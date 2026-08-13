# Security and Isolation Contract

Every document reaching this package is treated as hostile. The parser is
memory-unsafe C++ decoding an attacker-controlled binary format, and the
WebAssembly sandbox is the mechanism that contains it, not an incidental
implementation detail. Changes that weaken the boundary below require an
explicit decision recorded here.

## Isolation boundary

A conversion runs in a module instance created for that conversion alone and
closed afterwards. No instance is reused, so parser state cannot carry from one
document to the next.

The guest receives a read-only in-memory filesystem containing only the
document under conversion. It is granted no host filesystem access, no network,
and no environment variables. The document is never written to disk by this
package.

Memory unsafety in libcdr is therefore bounded by the guest's linear memory. A
corrupted parser can produce wrong output or fail; it cannot reach host memory,
the filesystem, or the network.

## Limits

Two limits are set explicitly, because the runtime defaults leave both open and
neither is safe for untrusted input.

Conversions honour context cancellation. The runtime is configured to terminate
guest execution when the context is cancelled or expires; without this a
malformed document can occupy a goroutine indefinitely and cancellation has no
effect.

Guest linear memory is capped at 512 MiB. Large documents legitimately need
room, but a crafted one must not be able to exhaust the host. Raising this
limit is a deliberate change, not a fix for a document that fails to convert.

SVG written to guest standard output is capped at 512 MiB before it reaches a
host buffer. A conversion beyond that limit fails with `ErrOutputTooLarge`.
Guest standard error is retained only up to 1 MiB for diagnostics; additional
bytes are discarded. These host-side limits prevent an otherwise sandboxed
guest from exhausting host memory through output streams.

## What is not guaranteed

The SVG this package returns is untrusted output derived from untrusted input.
SVG can carry script, event handlers, and external references. Callers that
render it in a browser, inline it into a page, or serve it from a trusted origin
must sanitise it first. Sanitisation is out of scope here and this package
performs none.

Conversion cost is not bounded beyond the limits above. A caller exposing this
to the public internet is responsible for request-level timeouts, concurrency
limits, and upload size limits appropriate to its deployment.
