# go-libcdr

[![CI](https://github.com/nobbs/go-libcdr/actions/workflows/ci.yaml/badge.svg)](https://github.com/nobbs/go-libcdr/actions/workflows/ci.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/nobbs/go-libcdr.svg)](https://pkg.go.dev/github.com/nobbs/go-libcdr)

`go-libcdr` converts CorelDRAW (`.cdr`) documents to SVG with
[libcdr](https://wiki.documentfoundation.org/DLP/Libraries/libcdr) compiled to
WebAssembly. It is a pure Go dependency: no cgo, external binary, or native
toolchain is needed by users of the package.

## Install

The package requires Go 1.25 or later.

```sh
go get github.com/nobbs/go-libcdr
```

## Convert a document

Create one converter and reuse it. Compiling the embedded WebAssembly module is
the expensive operation; `Converter` is safe for concurrent calls to
`Convert`. Call `Close` only after every conversion has returned.

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/nobbs/go-libcdr"
)

func main() {
    ctx := context.Background()

    converter, err := libcdr.New(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer func() { _ = converter.Close(context.Background()) }()

    document, err := os.ReadFile("drawing.cdr")
    if err != nil {
        log.Fatal(err)
    }

    svg, err := converter.Convert(ctx, document)
    if err != nil {
        log.Fatal(err)
    }

    if err := os.WriteFile("drawing.svg", svg, 0o644); err != nil {
        log.Fatal(err)
    }
}
```

The package converts the document's first page and returns libcdr's SVG
verbatim. In particular, CorelDRAW's preview bitmap is normally included as an
`<image>` element. Removing it, sanitising the SVG, or adjusting its bounds is
the caller's responsibility.

## Errors and limits

Use `errors.Is` to distinguish document rejection from conversion limits:

- `ErrUnsupportedDocument` means the input is empty or not recognised as a
  CorelDRAW document.
- `ErrParseFailed` means libcdr recognised the document but could not parse it.
- `ErrOutputTooLarge` means the SVG exceeded the 512 MiB output limit.

Cancelled or expired contexts are returned as wrapped context errors. Set a
deadline suitable for the service that calls the converter; document parsing
does not have a general execution-time bound.

Each conversion runs in a fresh WebAssembly instance with no host filesystem,
network, or environment access. Guest memory and SVG output are separately
capped at 512 MiB, and diagnostics are capped at 1 MiB. The resulting SVG is
still untrusted: sanitise it before rendering it in a browser or serving it
from a trusted origin.

## Development

The committed `cdr2svg.wasm` artifact is part of the module. Ordinary Go
development and tests do not need Docker or the WebAssembly build toolchain:

```sh
mise install
mise run check
```

Rebuild the module only when the documented build contract requires it. The
normative specifications cover the [conversion API](docs/specs/conversion.md),
[sandbox and limits](docs/specs/security.md), [native build](docs/specs/operations.md),
[distribution and licensing](docs/specs/distribution.md), and
[testing](docs/specs/testing.md).

## License

Original repository code is licensed under the [MIT License](LICENSE).
`cdr2svg.wasm` includes third-party code, including MPL-2.0-licensed libcdr and
librevenge. [NOTICE](NOTICE) identifies every bundled component, its licence,
and how to obtain its corresponding source.

CorelDRAW is a trademark of Corel Corporation. This project is not affiliated
with, endorsed by, or sponsored by Corel Corporation.
