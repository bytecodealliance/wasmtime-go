#!/bin/bash
set -euo pipefail

wasmtime=$1
if [ "$wasmtime" = "" ]; then
  echo "usage: $0 <path-to-wasmtime>"
  exit 1
fi

rm -rf build

# Build the C API with CMake (invokes cargo under the hood)
cmake -S "$wasmtime/crates/c-api" -B "$wasmtime/build-cmake" -DCMAKE_BUILD_TYPE=Release
cmake --build "$wasmtime/build-cmake"
rm -rf "$wasmtime/build-cmake"

# Create the expected directory structure with empty.go files
for d in "include" "include/wasmtime" "include/wasmtime/component" "include/wasmtime/component/types" "include/wasmtime/types" "linux-x86_64" "macos-x86_64" "windows-x86_64" "linux-aarch64" "macos-aarch64" "linux-riscv64"; do
  mkdir -p "build/$d"
  name=$(basename "$d" | tr - _)
  echo "package $name" > "build/$d/empty.go"
done

# Determine host platform
host_os=$(uname -s)
host_arch=$(uname -m)
case "$host_os-$host_arch" in
  Linux-x86_64)   platform=linux-x86_64; rust_target=x86_64-unknown-linux-gnu ;;
  Linux-aarch64)  platform=linux-aarch64; rust_target=aarch64-unknown-linux-gnu ;;
  Linux-riscv64)  platform=linux-riscv64; rust_target=riscv64gc-unknown-linux-gnu ;;
  Darwin-x86_64)  platform=macos-x86_64; rust_target=x86_64-apple-darwin ;;
  Darwin-arm64)   platform=macos-aarch64; rust_target=aarch64-apple-darwin ;;
  *) echo "Unsupported platform: $host_os-$host_arch"; exit 1 ;;
esac

# The CMake/cargo build puts artifacts in the wasmtime target directory.
# Native builds go to target/release/, cross builds to target/$rust_target/release/.
wasmtime_target_dir="$wasmtime/target/$rust_target/release"
if [ ! -d "$wasmtime_target_dir" ]; then
  wasmtime_target_dir="$wasmtime/target/release"
fi

# Find the built library (may be libwasmtime.a or wasmtime.lib)
built_lib=$(find "$wasmtime_target_dir" -maxdepth 1 -name "libwasmtime*.a" -o -name "wasmtime.lib" | head -1)
if [ -z "$built_lib" ]; then
  echo "Failed to find built wasmtime library in $wasmtime_target_dir"
  exit 1
fi
ln -s "$built_lib" "build/$platform/libwasmtime.a"

# Copy headers
cp "$wasmtime/crates/c-api/include/"*.h build/include/
cp -r "$wasmtime/crates/c-api/include/wasmtime" build/include/
