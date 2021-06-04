#!/bin/bash

wasmtime=$1
if [ "$wasmtime" = "" ]; then
  echo "must pass path to wasmtime"
  exit 1
fi

# Clean and re-create "build" directory hierarchy
rm -rf build
for d in "include" "include/wasmtime" "linux-x86_64" "macos-x86_64" "windows-x86_64"; do
  path="build/$d"
  mkdir -p "$path"
  name=$(basename $d)
  echo "package ${name/-/_}" > "$path/empty.go"
done

build="$wasmtime/target/release"
if [ ! -d "$build" ]; then
  build="$wasmtime/target/debug"
fi
# Use absolute path for symbolic links
build=$(readlink -f "$build")

if [ ! -f "$build/libwasmtime.a" ]; then
  echo 'Missing libwasmtime.a. Did you `cargo build -p wasmtime-c-api`?'
fi

ln -s "$build/libwasmtime.a" build/linux-x86_64/libwasmtime.a
ln -s "$build/libwasmtime.a" build/macos-x86_64/libwasmtime.a
cp "$wasmtime"/crates/c-api/include/*.h build/include
cp -r "$wasmtime"/crates/c-api/include/wasmtime build/include
cp "$wasmtime"/crates/c-api/wasm-c-api/include/*.h build/include
