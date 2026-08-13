# Testing Contract

Tests run without network access, without Docker, and without the WebAssembly
toolchain. They exercise the committed module, so a clone plus `go test ./...`
is sufficient.

## Fixtures

CorelDRAW documents are generally third-party artwork, so `testdata/` is not
committed and its contents are ignored by git. Tests that require a document
skip when the directory is empty and state why. Tests that do not require one
always run, so a fresh clone still verifies the package's behavior on bad input
and its cancellation contract.

Adding a fixture to CI requires a document whose redistribution rights are
unambiguous.

## Required coverage

Rejection tests prove that empty input, non-CorelDRAW bytes, a bare RIFF header,
and a truncated RIFF document each produce `ErrUnsupportedDocument` rather than
a panic, a hang, or a generic error. These tests are the guard on the exit-code
mapping between `native/cdr2svg.cpp` and the Go package; a change to either side
that breaks the mapping fails here.

Fixture tests prove that each document converts to output containing an `svg`
element, and that converting the same document twice yields byte-identical
results. Determinism is a stated part of the conversion contract and is
verified, not assumed.

A concurrency test converts one document from several goroutines against a
single `Converter` and proves every result matches the sequential result. The
package documents that a `Converter` is safe for concurrent use, so that claim
carries a test.

A cancellation test proves that a conversion started with a cancelled context
fails with `context.Canceled` rather than succeeding or blocking.

Unit tests prove that output buffering rejects writes above its configured cap
and that diagnostic buffering retains only its configured prefix. These tests
cover the host-memory limits separately from the guest's WebAssembly memory
limit.

## Commit-time guards

`prek` runs the hooks in `.pre-commit-config.yaml` before each commit. Alongside
the generic file and Markdown checks, two are specific to this repository:
`mise run fix` formats and applies lint fixes to changed Go files, and
`native/build/check-exit-codes.sh` fails the commit when the exit codes in
`native/cdr2svg.cpp` no longer match the constants in `libcdr.go`. That guard
exists because the two sides are a single contract that nothing else enforces.

## Verifying a module rebuild

A rebuild is checked against a native build of `native/cdr2svg.cpp` by comparing
SHA-256 digests of the SVG produced from the same fixtures. This is a manual
step performed when the module changes, not part of `go test`, because it
requires the toolchain image.
