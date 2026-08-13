// Package libcdr converts CorelDRAW documents to SVG.
//
// Parsing is done by libcdr compiled to WebAssembly and executed with wazero,
// so this package is pure Go: no cgo, no external binary, nothing to install
// alongside it.
//
// The parser is memory-unsafe C++ consuming untrusted input, and the
// WebAssembly sandbox is what contains that: each conversion gets a fresh
// module instance, a read-only in-memory filesystem holding only the input, no
// network, and no host filesystem access. Conversions also honour context
// cancellation and run under a memory ceiling.
//
// Output is exactly what libcdr produces. In particular, CorelDRAW's embedded
// preview bitmap is passed through as an <image> element, which is typically
// the bulk of the bytes; removing it is the caller's decision, not a conversion
// step.
package libcdr

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"testing/fstest"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
)

//go:embed cdr2svg.wasm
var converterModule []byte

// Exit codes reported by native/cdr2svg.cpp. Keep both sides in sync.
const (
	exitUnsupported = 3
	exitParseFailed = 4
)

// guestInputPath is where the document is mounted inside the sandbox.
const guestInputPath = "input.cdr"

// memoryLimitPages caps the guest's linear memory at 512 MiB (8192 * 64 KiB).
// Large documents legitimately need room, but a crafted one must not be able to
// exhaust the host.
const memoryLimitPages = 8192

const (
	// maxOutputBytes keeps untrusted guest output from allocating unbounded host
	// memory. It matches the guest memory limit: a successful conversion can
	// return at most as much SVG as the guest can hold at once.
	maxOutputBytes = memoryLimitPages * 64 * 1024

	// maxDiagnosticBytes retains enough stderr for errors without allowing an
	// untrusted guest to grow a host buffer through diagnostics.
	maxDiagnosticBytes = 1 << 20
)

// ErrUnsupportedDocument reports that the input is not a CorelDRAW document
// libcdr recognises. It is distinct from a parse failure so that callers can
// answer "wrong file type" differently from "this file broke the parser".
var ErrUnsupportedDocument = errors.New("libcdr: unsupported document")

// ErrParseFailed reports that the input was recognised as a CorelDRAW document
// but could not be parsed.
var ErrParseFailed = errors.New("libcdr: parsing the document failed")

// ErrOutputTooLarge reports that the converter produced more SVG than this
// package's output limit permits.
var ErrOutputTooLarge = errors.New("libcdr: conversion output exceeds the size limit")

// Converter converts CorelDRAW documents to SVG.
//
// Compiling the WebAssembly module is the expensive step, so create one
// Converter and reuse it. The zero value is not usable; call [New]. A Converter
// is safe for concurrent calls to [Converter.Convert]. Call [Converter.Close]
// only after those calls have returned.
type Converter struct {
	runtime  wazero.Runtime
	compiled wazero.CompiledModule
}

// New compiles the embedded converter module. Call [Converter.Close] when done.
func New(ctx context.Context) (*Converter, error) {
	config := wazero.NewRuntimeConfig().
		// libcdr signals parse errors with C++ exceptions, which compile to the
		// WebAssembly exception-handling proposal.
		WithCoreFeatures(api.CoreFeaturesV2 | experimental.CoreFeaturesExceptionHandling).
		// Without this, a malformed document can spin forever and cancelling
		// the context will not stop it.
		WithCloseOnContextDone(true).
		WithMemoryLimitPages(memoryLimitPages)

	runtime := wazero.NewRuntimeWithConfig(ctx, config)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, runtime); err != nil {
		_ = runtime.Close(ctx)
		return nil, fmt.Errorf("instantiating wasi: %w", err)
	}

	compiled, err := runtime.CompileModule(ctx, converterModule)
	if err != nil {
		_ = runtime.Close(ctx)
		return nil, fmt.Errorf("compiling the converter module: %w", err)
	}

	return &Converter{runtime: runtime, compiled: compiled}, nil
}

// Close releases the compiled module and its runtime. It must not run while
// [Converter.Convert] is in progress.
func (c *Converter) Close(ctx context.Context) error {
	return c.runtime.Close(ctx)
}

// Convert parses a CorelDRAW document and returns the SVG for its first page.
//
// The document is never written to disk. Callers can distinguish rejection
// causes with errors.Is against [ErrUnsupportedDocument], [ErrParseFailed],
// and [ErrOutputTooLarge].
func (c *Converter) Convert(ctx context.Context, document []byte) ([]byte, error) {
	if len(document) == 0 {
		return nil, fmt.Errorf("%w: empty input", ErrUnsupportedDocument)
	}

	input := fstest.MapFS{
		guestInputPath: &fstest.MapFile{Data: document, Mode: 0o444},
	}

	stdout := limitedBuffer{limit: maxOutputBytes}
	stderr := truncatedBuffer{limit: maxDiagnosticBytes}

	config := wazero.NewModuleConfig().
		WithName(""). // anonymous, so concurrent conversions do not collide
		WithArgs("cdr2svg", "/"+guestInputPath).
		WithStdout(&stdout).
		WithStderr(&stderr).
		WithFSConfig(wazero.NewFSConfig().WithFSMount(fs.FS(input), "/"))

	module, err := c.runtime.InstantiateModule(ctx, c.compiled, config)
	if module != nil {
		defer func() { _ = module.Close(ctx) }()
	}

	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, fmt.Errorf("libcdr: conversion aborted: %w", ctxErr)
	}

	if stdout.overflowed {
		return nil, fmt.Errorf("%w: limit is %d bytes", ErrOutputTooLarge, maxOutputBytes)
	}

	if err != nil {
		return nil, convertError(ctx, err, stderr.Bytes())
	}

	if stdout.buffer.Len() == 0 {
		return nil, fmt.Errorf("%w: converter produced no output: %s",
			ErrParseFailed, bytes.TrimSpace(stderr.Bytes()))
	}

	return stdout.buffer.Bytes(), nil
}

// convertError maps a module failure onto the package's sentinel errors.
func convertError(ctx context.Context, err error, stderr []byte) error {
	// A cancelled or expired context surfaces as an exit error once
	// WithCloseOnContextDone terminates the guest; report the cause.
	if ctxErr := ctx.Err(); ctxErr != nil {
		return fmt.Errorf("libcdr: conversion aborted: %w", ctxErr)
	}

	var exit *sys.ExitError
	if errors.As(err, &exit) {
		detail := bytes.TrimSpace(stderr)

		switch exit.ExitCode() {
		case exitUnsupported:
			return fmt.Errorf("%w: %s", ErrUnsupportedDocument, detail)
		case exitParseFailed:
			return fmt.Errorf("%w: %s", ErrParseFailed, detail)
		default:
			return fmt.Errorf("libcdr: converter exited with code %d: %s",
				exit.ExitCode(), detail)
		}
	}

	return fmt.Errorf("libcdr: running the converter: %w", err)
}

type limitedBuffer struct {
	buffer     bytes.Buffer
	limit      int
	overflowed bool
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.buffer.Len() {
		b.overflowed = true
		return 0, ErrOutputTooLarge
	}

	return b.buffer.Write(p)
}

type truncatedBuffer struct {
	buffer bytes.Buffer
	limit  int
}

func (b *truncatedBuffer) Write(p []byte) (int, error) {
	n := len(p)

	remaining := b.limit - b.buffer.Len()
	if remaining > 0 {
		if len(p) > remaining {
			p = p[:remaining]
		}

		if _, err := b.buffer.Write(p); err != nil {
			return 0, err
		}
	}

	return n, nil
}

func (b *truncatedBuffer) Bytes() []byte {
	return b.buffer.Bytes()
}
