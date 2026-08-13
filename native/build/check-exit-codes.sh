#!/bin/sh
# The native converter's exit codes are part of the Go package's error taxonomy
# (docs/specs/conversion.md). This fails the commit when the two drift apart.
set -eu

cd "$(dirname "$0")/../.."

value_from_cpp() {
  sed -n "s/.*$1 = \([0-9][0-9]*\).*/\1/p" native/cdr2svg.cpp | head -1
}

value_from_go() {
  sed -n "s/.*$1 = \([0-9][0-9]*\).*/\1/p" libcdr.go | head -1
}

status=0

check() {
  name=$1
  cpp_symbol=$2
  go_symbol=$3

  cpp_value=$(value_from_cpp "$cpp_symbol")
  go_value=$(value_from_go "$go_symbol")

  if [ -z "$cpp_value" ] || [ -z "$go_value" ]; then
    echo "check-exit-codes: could not read $name ($cpp_symbol / $go_symbol)" >&2
    status=1
    return
  fi

  if [ "$cpp_value" != "$go_value" ]; then
    echo "check-exit-codes: $name differs: $cpp_symbol=$cpp_value but $go_symbol=$go_value" >&2
    status=1
  fi
}

check "unsupported document" EXIT_UNSUPPORTED exitUnsupported
check "parse failure" EXIT_PARSE_FAILED exitParseFailed

exit $status
