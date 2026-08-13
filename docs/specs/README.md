# Specifications

This directory contains the normative behavior contracts for `go-libcdr`.

- [`conversion.md`](conversion.md) defines the conversion API, output
  semantics, determinism, and the error taxonomy.
- [`security.md`](security.md) defines the isolation boundary for untrusted
  documents and the limits placed on a conversion.
- [`operations.md`](operations.md) defines the WebAssembly build contract: the
  pinned toolchain, the compile and post-processing flags, and when the
  committed module is rebuilt.
- [`distribution.md`](distribution.md) defines the module path, the embedded
  artifact, versioning, and the obligations that come with redistributing
  MPL-2.0 code.
- [`testing.md`](testing.md) defines the required automated verification.

These specifications describe intended behavior. They do not silently change to
match an implementation defect, an upstream libcdr change, or a temporary
exception. Resolve a conflict explicitly and update the relevant specification
together with the implementation and tests.
