#!/bin/sh

if [ "$WASMTIME" = "" ]; then
  echo "must set \$WASMTIME env var"
  exit 1
fi

build=$WASMTIME/target/release
if [ ! -d $build ]; then
  build=$WASMTIME/target/debug
fi

export CGO_LDFLAGS="-L$build -Wl,-rpath,$build"
export CGO_CFLAGS="-I$WASMTIME/crates/c-api/wasm-c-api/include -I$WASMTIME/crates/c-api/include"

exec "$@"
