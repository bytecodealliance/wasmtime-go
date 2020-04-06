#!/bin/sh

if [ "$WASMTIME" = "" ]; then
  echo "must set \$WASMTIME env var"
  exit 1
fi

rm -rf pkg/wasmtime/build

build=$WASMTIME/target/release
if [ ! -d $build ]; then
  build=$WASMTIME/target/debug
fi

mkdir -p pkg/wasmtime/build/include
ln -s $build pkg/wasmtime/build/lib
cp $WASMTIME/crates/c-api/include/*.h pkg/wasmtime/build/include
cp $WASMTIME/crates/c-api/wasm-c-api/include/*.h pkg/wasmtime/build/include
