# Repository Guidance

Read [`CONTRIBUTING.md`](CONTRIBUTING.md) first: it defines setup, the required
`mise run check` gate, fixture handling, pull-request conventions, and the
signed-commit and linear-history rules. This file adds only what an agent needs
on top of that.

## Working Model

Complete work as a single agent by default. Do not delegate specification
review, test coverage, or validation to subagents. Preserve the quality bar
through the checks below and report the evidence directly.

Keep changes scoped to the request. Preserve unrelated working-tree changes,
stage only intended files, and do not weaken a check, widen a limit, or add a
lint suppression to make a change pass. If a check is genuinely wrong, say so
and fix the check deliberately.

This package runs a memory-unsafe C++ parser against untrusted input inside a
WebAssembly sandbox. Treat the isolation boundary, the resource limits, the
error taxonomy, and the licensing obligations of the committed module as
contracts rather than implementation details.

## Which Specification Applies

`docs/specs/` is normative, and `CONTRIBUTING.md` requires updating it in the
same change. Use the specification that matches the change rather than reading
all of them:

- the conversion API, output semantics, determinism, and errors:
  `conversion.md`;
- the sandbox boundary, memory and output limits, and untrusted output:
  `security.md`;
- the WebAssembly toolchain, compile flags, and rebuild triggers:
  `operations.md`;
- the module path, Go version floor, embedded artifact, versioning, and
  MPL-2.0 obligations: `distribution.md`; and
- expected test coverage: `testing.md`.

A specification and the code disagreeing is a bug in one of them. Resolve it
explicitly; do not quietly reword the specification to match what the code
happens to do.

## Invariants Worth Knowing

These are easy to break and cheap to respect:

- **The package is flat.** It is a library: `internal/` is not used, and
  layer-named packages are not introduced.
- **Exit codes are one contract across two languages.** The values in
  `native/cdr2svg.cpp` and the constants in `libcdr.go` must agree;
  `native/build/check-exit-codes.sh` runs as a commit hook and fails when they
  do not. Fix both sides, never one.
- **Limits are deliberate.** `memoryLimitPages` and `maxOutputBytes` bound what
  a hostile document can cost the host. Raising either is a decision recorded in
  `security.md`, not a fix for a document that fails to convert.
- **The Go floor is the oldest version that compiles**, not the newest
  available. Raising it taxes every dependent.
- **`cdr2svg.wasm` is a committed artifact.** Go-only changes never rebuild it.
  Every rebuild adds a permanent binary diff to history.

## Reporting

Run `mise run check` before reporting completion and report what actually
passed. Never describe an unrun check as passing, and state plainly when
something is skipped: fixture tests skip on a clean checkout because
`testdata/` is empty, so a green run does not by itself prove that conversion
of a real document still works.
