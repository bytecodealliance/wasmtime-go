#!/bin/bash

wasmtime=$1
target=$2
if [ "$wasmtime" = "" ]; then
  echo "must pass path to wasmtime"
  exit 1
fi

# Clean and re-create "build" directory hierarchy
rm -rf build
for d in "include" "include/wasmtime" "linux-x86_64" "macos-x86_64" "windows-x86_64" "linux-aarch64" "macos-aarch64"; do
  path="build/$d"
  mkdir -p "$path"
  name=$(basename $d)
  echo "package ${name/-/_}" > "$path/empty.go"
done

build="$wasmtime/target/"$target"/release"
if [ ! -d "$build" ]; then
  build="$wasmtime/target/"$target"/debug"
fi
# Use absolute path for symbolic links
build=$(cd "$build" && pwd)

if [ ! -f "$build/libwasmtime.a" ]; then
  echo 'Missing libwasmtime.a. Did you `cargo build -p wasmtime-c-api`?'
fi

for d in "linux-x86_64" "macos-x86_64" "linux-aarch64" "macos-aarch64"; do
  ln -s "$build/libwasmtime.a" "build/$d/libwasmtime.a"
done

cp "$wasmtime"/crates/c-api/include/*.h build/include
cp -r "$wasmtime"/crates/c-api/include/wasmtime build/include
cp "$wasmtime"/crates/c-api/wasm-c-api/include/*.h build/include
