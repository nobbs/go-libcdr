#!/bin/sh
# Download the upstream sources compiled into cdr2svg.wasm.
# They are deliberately not vendored; run this once before build.sh.
set -eu

cd "$(dirname "$0")"

fetch() {
  url=$1
  archive=$2
  directory=$3
  checksum=$4
  if [ -d "$directory" ]; then
    echo "have $directory"
    return
  fi

  echo "fetching $directory"
  curl -fsSL --retry 5 --retry-delay 3 --retry-all-errors "$url" -o "$archive"
  if command -v sha256sum >/dev/null 2>&1; then
    printf '%s  %s\n' "$checksum" "$archive" | sha256sum -c -
  elif command -v shasum >/dev/null 2>&1; then
    printf '%s  %s\n' "$checksum" "$archive" | shasum -a 256 -c -
  else
    echo "fetch: need sha256sum or shasum to verify $archive" >&2
    exit 1
  fi
  tar xf "$archive"
  rm -f "$archive"
}

fetch https://dev-www.libreoffice.org/src/libcdr/libcdr-0.1.7.tar.bz2 \
      libcdr.tar.bz2 libcdr-0.1.7 \
      ae613caeb7e0e539cbc1b08ea5169bddaed8d2021d25ef66b39ddc0aa72c2902
fetch https://dev-www.libreoffice.org/src/librevenge-0.0.6.tar.bz2 \
      librevenge.tar.bz2 librevenge-0.0.6 \
      52a65e904d255dbdd97a8b7bb28d6574e14f999eb01416aff004502406d0904d
fetch https://dev-www.libreoffice.org/src/lcms2-2.12.tar.gz \
      lcms2.tar.gz lcms2-2.12 \
      18663985e864100455ac3e507625c438c3710354d85e5cbb7cd4043e11fe10f5
fetch https://github.com/madler/zlib/releases/download/v1.3.1/zlib-1.3.1.tar.gz \
      zlib.tar.gz zlib-1.3.1 \
      9a93b2b7dfdac77ceba5a558a580e74667dd6fede4585b91eefb60f03b72df23

echo "sources ready"
