// Minimal CorelDRAW -> SVG: libcdr parses, librevenge emits. No Inkscape.
//
// Exit codes are distinct so the calling Go package can tell "this is not a
// CorelDRAW file" (answer 415) from "it is one and parsing failed" (answer 500).
// Keep these in sync with the constants in libcdr.go.
#include <librevenge/librevenge.h>
#include <librevenge-generators/librevenge-generators.h>
#include <librevenge-stream/librevenge-stream.h>
#include <libcdr/libcdr.h>
#include <iostream>

enum ExitCode
{
  EXIT_OK = 0,
  EXIT_USAGE = 2,
  EXIT_UNSUPPORTED = 3,
  EXIT_PARSE_FAILED = 4
};

int main(int argc, char **argv)
{
  if (argc < 2)
  {
    std::cerr << "usage: cdr2svg FILE.cdr\n";
    return EXIT_USAGE;
  }

  librevenge::RVNGFileStream input(argv[1]);
  if (!libcdr::CDRDocument::isSupported(&input))
  {
    std::cerr << "not a supported CorelDRAW document\n";
    return EXIT_UNSUPPORTED;
  }

  librevenge::RVNGStringVector output;
  librevenge::RVNGSVGDrawingGenerator generator(output, "");
  if (!libcdr::CDRDocument::parse(&input, &generator))
  {
    std::cerr << "parsing the document failed\n";
    return EXIT_PARSE_FAILED;
  }

  if (output.empty())
  {
    std::cerr << "converter produced no output\n";
    return EXIT_PARSE_FAILED;
  }

  // librevenge returns one SVG per document page. The Go API exposes the
  // first page, so emitting any later page would concatenate XML documents.
  std::cout << "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" << output[0].cstr() << "\n";
  return EXIT_OK;
}
