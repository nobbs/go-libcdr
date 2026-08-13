# Security Policy

## Supported version

No version has been released yet. The current `main` branch is the supported
development version.

## Reporting a vulnerability

Please do not report security vulnerabilities in a public issue. Instead,
[privately report a vulnerability on GitHub](https://github.com/nobbs/go-libcdr/security/advisories/new).
Include the affected revision, a minimal reproduction, the impact you observed,
and any relevant environment details. Do not include sensitive documents unless
they are necessary to reproduce the issue and you have the right to share them.

## Scope

This policy covers the Go package, the committed `cdr2svg.wasm` module, and the
native build and GitHub Actions configuration in this repository.

The package accepts untrusted CorelDRAW documents and executes the memory-unsafe
libcdr parser inside WebAssembly. Reports are especially useful when they show
that a document can:

- access the host filesystem, network, or environment;
- escape the WebAssembly sandbox or affect host memory;
- bypass guest-memory, SVG-output, or diagnostic-output limits;
- ignore context cancellation or otherwise cause unbounded package-level work;
- cause the package to expose data from another conversion; or
- produce SVG that is treated as trusted by this package.

The package returns untrusted SVG and deliberately does not sanitise it.
Applications that render, inline, or serve the SVG must apply their own
sanitisation, origin isolation, request timeouts, upload limits, and concurrency
limits. A vulnerability in those caller-controlled protections is outside this
repository's scope unless it bypasses a guarantee made by this package.

Malformed or unsupported documents that fail safely are not vulnerabilities by
themselves. Upstream vulnerabilities in bundled dependencies are in scope when
they are reachable through the committed module or its build inputs.

See the normative [security and isolation contract](docs/specs/security.md) for
the package's documented boundary and limits.
