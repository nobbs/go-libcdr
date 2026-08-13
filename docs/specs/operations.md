# Build and Operations Contract

`cdr2svg.wasm` is the intended committed build artifact. This specification
defines its production so that the source inputs and toolchain are independently
verifiable.

## Pinned inputs

The build compiles libcdr 0.1.7, librevenge 0.0.6, lcms2 2.12, and zlib 1.3.1.
`native/build/fetch.sh` records each archive URL and SHA-256 digest, verifies it
before extraction, and does not vendor the source trees. No upstream source is
patched; adaptations belong in `native/shim/` or compiler flags.

`native/build/Dockerfile.wasi` pins the Debian base image by digest, wasi-sdk 33,
and Binaryen 131. Its wasi-sdk and Binaryen archives are SHA-256 verified for
both supported builder architectures: arm64 and amd64.

## Build procedure

Run the fetcher from the repository root, then build the toolchain image and
run the compiler inside it:

```sh
sh native/build/fetch.sh
docker build -f native/build/Dockerfile.wasi -t go-libcdr-wasi:33 .
docker run --rm -v "$PWD":/work go-libcdr-wasi:33 sh native/build/build.sh
```

`build.sh` compiles directly for `wasm32-wasip1` without autotools. The result
is written to `cdr2svg.wasm` at the repository root so the future Go package can
embed it without an install-time build.

## Compilation requirements

Exceptions and RTTI are required: libcdr uses exceptions for parse control
flow, and librevenge/libcdr use `dynamic_cast`. The build enables Wasm
exceptions, links the `eh` sysroot variant and `-lunwind`, and keeps RTTI on.

`-DBOOST_NO_CXX98_FUNCTION_BASE` avoids Boost 1.74's removed
`std::unary_function` dependency. `-Wno-register` is required for lcms2 2.12
headers.

Binaryen translates LLVM's legacy exception encoding to exnref, which wazero
supports, then removes debug and producer sections. This post-processing is
part of the build contract.

## Rebuild triggers and verification

Rebuild the module when a pinned upstream version, toolchain pin, compiler or
post-processing flag, `native/cdr2svg.cpp`, or a native shim changes. Do not
rebuild for Go-only changes.

Every module rebuild must convert the available fixtures and compare the SVG
output with a native build of the same converter source. A difference is a
regression in the Wasm build, not an acceptable platform variation.
