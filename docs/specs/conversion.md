# Conversion Contract

The package converts a CorelDRAW document, supplied as bytes, into the SVG that
libcdr produces for that document's first page. Conversion is a translation
step and nothing more: the package adds no cleanup, no normalization, and no
interpretation of its own.

## Entry point and lifetime

`Converter` is the entry point. Compiling the embedded WebAssembly module is
the expensive part of the work, so a `Converter` is created once with `New` and
reused; the zero value is not usable. `Close` releases the compiled module and
its runtime.

A `Converter` is safe for concurrent calls to `Convert`. Each conversion runs
in its own module instance, and no state carries from one conversion to the
next. `Close` must run only after all conversions have returned.

`Convert` accepts the whole document as a byte slice rather than a reader. The
sandboxed guest is given a filesystem holding the document, so the entire input
is required before conversion starts; the signature states that cost honestly.

## Output

Output is libcdr's, verbatim. Two consequences are part of the contract rather
than defects:

The SVG's page is the CorelDRAW document page, not the bounding box of the
artwork. A drawing may sit anywhere on that page, including outside it.

CorelDRAW's embedded preview bitmap is carried through as an `<image>` element.
It is frequently the overwhelming majority of the output. Removing it, and
recomputing a bounding box afterwards, belongs to the caller.

Conversion is deterministic. The same input always produces byte-identical
output for a given build of the module.

## Errors

Callers can distinguish why a document was rejected, because the two cases
warrant different responses in a service:

`ErrUnsupportedDocument` reports input libcdr does not recognise as a CorelDRAW
document. Empty input is reported the same way.

`ErrParseFailed` reports input that was recognised but could not be read, and
also covers a conversion that completes without producing output.

`ErrOutputTooLarge` reports SVG that exceeds the package's 512 MiB output
limit. This is distinct from a parser failure: callers may choose to reject the
document, reduce their requested output, or use another conversion path.

A cancelled or expired context is reported as a wrapped `context` error, not as
a parse failure.

The native converter distinguishes these with distinct exit codes — 3 for
unsupported, 4 for parse failure — and the Go package maps them. Both sides
must change together; neither may rely on the text written to stderr.

## Text encoding

libcdr uses ICU only to transcode text runs from legacy CorelDRAW codepages,
and the build replaces ICU with a minimal shim. UTF-16LE and windows-1252 are
decoded faithfully. Other legacy codepages fall back to windows-1252, so text
runs in those encodings may be transcoded incorrectly. Geometry, colour, and
layers are unaffected, and no other part of the output depends on ICU.
