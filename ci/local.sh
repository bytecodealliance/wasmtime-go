#!/bin/bash

if [ "$WASMTIME" = "" ]; then
  echo "must set \$WASMTIME env var"
  exit 1
fi

rm -rf pkg/wasmtime/build

build=$WASMTIME/target/release
if [ ! -d $build ]; then
  build=$WASMTIME/target/debug
fi

mkdir -p pkg/wasmtime/build/{include,linux-x86_64,macos-x86_64}
ln -s $build/libwasmtime.a pkg/wasmtime/build/linux-x86_64/libwasmtime.a
ln -s $build/libwasmtime.a pkg/wasmtime/build/macos-x86_64/libwasmtime.a
cp $WASMTIME/crates/c-api/include/*.h pkg/wasmtime/build/include
cp $WASMTIME/crates/c-api/wasm-c-api/include/*.h pkg/wasmtime/build/include
