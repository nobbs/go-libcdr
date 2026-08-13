# Distribution and Licensing Contract

## Module path

When the Go package is introduced, its module path will be
`github.com/nobbs/go-libcdr` and its package name will be `libcdr`. The path is
permanent: it is recorded in the `go.mod` of every dependent, and the module
proxy caches published versions immutably. Changing it is not a rename; it
abandons the module and starts another.

The package requires Go 1.26 or later. `go.mod` suggests the current Go 1.26.5
patch toolchain for development and reproducible checks; supported older patch
releases remain a Go distribution decision rather than an application fallback.

## The embedded artifact

`cdr2svg.wasm` is committed at the repository root so the forthcoming Go package
can embed it with `go:embed`. The package will not fetch or build it at install
time, `go generate` time, or runtime: each would require every dependent to
install a toolchain, which defeats the package's reason to exist.

Git LFS must not be used for it. The module proxy fetches without LFS
smudging, so dependents would receive a pointer file.

The module is kept small deliberately — stripping is part of the build contract
in [`operations.md`](operations.md) — and is rebuilt only when that contract
says so.

## Versioning

When release automation is introduced, versions will be derived from commit
history by Release Please. It will read Conventional Commits and open a release
pull request carrying the version bump and changelog entry. Merging that pull
request will publish the tag; nothing is tagged by hand.

`feat` produces a minor bump, `fix` and `deps` a patch. While the module is
below `v1`, a breaking change is a minor bump rather than a major one, so the
API is not yet stable by semantic-versioning convention.

Because pull requests are squash-merged, the pull request title becomes the
commit subject Release Please reads. A title that is not a valid Conventional
Commit drops its change out of the release history silently rather than
failing, so titles follow the Conventional Commits types listed above.

Release Please will run on every push to the default branch, covering both
opening the release pull request and cutting the release when that pull request
is merged. Applying the `release-please:force-run` label to a pull request will
re-run it against that pull request's base branch as a manual recovery path when
a release pull request has gone stale.

No module version is published by this change. Once publication is enabled, tags
will be created only from a state that passes the repository's required checks.
A published version cannot be withdrawn from the module proxy, so the first tag
waits until the module path and package API are settled.

## Third-party code

`cdr2svg.wasm` contains compiled libcdr and librevenge, both under MPL-2.0.
Distributing the module therefore distributes MPL-2.0 code in Executable Form,
and MPL-2.0 section 3.2 obliges this repository to tell recipients how to obtain
the corresponding Source Code Form.

`NOTICE` discharges that obligation: it names every bundled project, its version,
its licence, and the exact URL its source is fetched from. It also includes the
complete notices required by lcms2 and zlib; the ICU-derived compatibility macro
retains its full notice in `native/shim/NOTICE`. `native/build/fetch.sh` pins the
same source archives and `native/build/build.sh` reproduces the module from them.
`NOTICE` and those build inputs change together.

MPL-2.0 is file-level copyleft and no bundled file is modified, so it does not
extend to original work in this repository, which is MIT licensed. Where
librevenge is offered under MPL-2.0 or LGPL-2.1+, this repository distributes it
under MPL-2.0.

CorelDRAW is a trademark of Corel Corporation. This project is not affiliated
with, endorsed by, or sponsored by Corel Corporation, and neither the module
path nor the package name uses the mark.
