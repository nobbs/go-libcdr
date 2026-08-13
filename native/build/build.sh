#!/bin/sh
# Compile the fetched native sources into a standalone wasm32-wasip1 module.
set -eu

ROOT=/work
SRC=$ROOT/native/build
OBJ=/tmp/cdr2svg-obj
TARGET=wasm32-wasip1
EH_LIBS=/opt/wasi-sdk/share/wasi-sysroot/lib/$TARGET/eh

LIBCDR=$SRC/libcdr-0.1.7
LIBREVENGE=$SRC/librevenge-0.0.6
LCMS2=$SRC/lcms2-2.12
ZLIB=$SRC/zlib-1.3.1

for dependency in "$LIBCDR" "$LIBREVENGE" "$LCMS2" "$ZLIB"; do
  [ -d "$dependency" ] || {
    echo "missing $dependency; run native/build/fetch.sh first" >&2
    exit 1
  }
done

rm -rf "$OBJ"
mkdir -p "$OBJ"

INCLUDES="-I$ROOT/native/shim -I/opt/boost/include \
-I$LIBREVENGE/inc -I$LIBREVENGE/src/lib \
-I$LIBCDR/inc -I$LIBCDR/src/lib \
-I$LCMS2/include -I$ZLIB"

# Boost 1.74 references std::unary_function, removed in C++17. lcms2 uses
# register in its headers. RTTI and exceptions are required by librevenge and
# libcdr respectively.
CXXFLAGS="--target=$TARGET -O2 -fwasm-exceptions -std=c++17 -DNDEBUG \
-DBOOST_NO_CXX98_FUNCTION_BASE \
-Wno-deprecated-builtins -Wno-deprecated-declarations -Wno-register $INCLUDES"
CFLAGS="--target=$TARGET -O2 -DNDEBUG -Wno-register $INCLUDES"

compile_c() {
  for file in "$@"; do
    clang $CFLAGS -c "$file" -o "$OBJ/$(echo "$file" | tr '/' '_').o"
  done
}

compile_cxx() {
  for file in "$@"; do
    clang++ $CXXFLAGS -c "$file" -o "$OBJ/$(echo "$file" | tr '/' '_').o"
  done
}

echo "==> zlib"
compile_c "$ZLIB"/adler32.c "$ZLIB"/crc32.c "$ZLIB"/inffast.c "$ZLIB"/inflate.c \
          "$ZLIB"/inftrees.c "$ZLIB"/zutil.c "$ZLIB"/uncompr.c

echo "==> lcms2"
compile_c "$LCMS2"/src/*.c

echo "==> librevenge"
compile_cxx "$LIBREVENGE"/src/lib/*.cpp

echo "==> libcdr"
compile_cxx "$LIBCDR"/src/lib/*.cpp

echo "==> ICU shim and converter"
compile_cxx "$ROOT"/native/shim/icu_shim.cpp "$ROOT"/native/cdr2svg.cpp

echo "==> link"
clang++ --target=$TARGET -O2 -fwasm-exceptions -L"$EH_LIBS" -lunwind \
  -o /tmp/cdr2svg.raw.wasm "$OBJ"/*.o

# Convert LLVM's legacy exception representation to the exnref proposal used
# by wazero, then discard non-runtime debug and producer sections.
echo "==> translate exceptions and strip"
wasm-opt --enable-exception-handling --enable-bulk-memory --translate-to-exnref \
  --strip-debug --strip-producers -O2 \
  /tmp/cdr2svg.raw.wasm -o "$ROOT"/cdr2svg.wasm
rm -f /tmp/cdr2svg.raw.wasm

ls -l "$ROOT"/cdr2svg.wasm
