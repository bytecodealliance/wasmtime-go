#!/bin/bash

wasmtime=$1
if [ "$wasmtime" = "" ]; then
  echo "must pass path to wasmtime"
  exit 1
fi

rm -rf build

build=$wasmtime/target/release
if [ ! -d $build ]; then
  build=$wasmtime/target/debug
fi

if [ ! -f $build/libwasmtime.a ]; then
  echo 'Missing libwasmtime.a. Did you `cargo build -p wasmtime-c-api`?'
fi

mkdir -p build/{include,linux-x86_64,macos-x86_64}
ln -s $build/libwasmtime.a build/linux-x86_64/libwasmtime.a
ln -s $build/libwasmtime.a build/macos-x86_64/libwasmtime.a
cp $wasmtime/crates/c-api/include/*.h build/include
cp $wasmtime/crates/c-api/wasm-c-api/include/*.h build/include
