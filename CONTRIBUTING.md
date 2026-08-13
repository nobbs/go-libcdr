# Contributing

Contributions that keep the package small, deterministic, and safe to use with
untrusted documents are welcome. Please open an issue before starting a larger
change so that its contract and scope can be agreed first.

## Set up

The project pins Go and its development tools with [Mise](https://mise.jdx.dev/).
From the repository root:

```sh
mise install
mise run hooks-install
mise run check
```

`mise run check` is the required local gate. It runs linting, a tidy check,
normal and race-tested behavior, vulnerability scanning, and the commit hooks.
The ordinary test suite does not require network access, Docker, or the
WebAssembly toolchain.

## Tests and fixtures

Tests that do not need a CorelDRAW document always run. To exercise fixture
tests locally, place `.cdr` files in `testdata/`; they are ignored by Git.
Only add a fixture to the repository or CI when its redistribution rights are
unambiguous, such as a document you authored or one released under CC0.

## Change contracts

The specifications in [`docs/specs/`](docs/specs/README.md) are normative.
Update the relevant specification in the same change whenever you alter the
conversion API or errors, sandbox boundary or limits, native build contract, or
distribution and licensing obligations.

The WebAssembly module is a committed build artifact. Do not rebuild it for
Go-only changes. Native changes require the rebuild triggers and fixture
comparison in [`operations.md`](docs/specs/operations.md); do not edit the
downloaded upstream sources under `native/build/`. Express an accommodation as
a compiler flag or an addition under `native/shim/` instead.

## Pull requests

Keep each pull request focused and include tests for changed behavior and error
paths. Pull-request titles use Conventional Commits because Release Please uses
the eventual squash-merge subject to determine versions and changelog entries.

Use an allowed type such as `feat`, `fix`, `deps`, `docs`, `test`, `build`, or
`ci`; use `feat` for user-visible functionality, `fix` for corrections, and
`deps` for dependency updates. See the [versioning contract](docs/specs/distribution.md)
for how those types affect releases.

Please report vulnerabilities privately under the [security policy](SECURITY.md),
not in a public issue.
