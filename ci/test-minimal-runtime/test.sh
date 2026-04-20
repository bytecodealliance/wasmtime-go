#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
trap 'rm -rf vendor module.cwasm' EXIT

# Step 1: Create a pre-compiled module using the full Wasmtime library.
go run create_cwasm.go

# Step 2: Vendor the module, then adjust the vendored copy so it compiles
# against the minimal Wasmtime library:
#   a) Remove Go source files that call C functions absent from the min binary.
#   b) Download the min static libraries over the full ones so CGO links the right lib.
DOWNLOAD_SCRIPT="$(pwd)/../download-wasmtime.py"
go mod vendor
(
  cd vendor/github.com/bytecodealliance/wasmtime-go/v44
  rm -f wat2wasm.go wasi.go *_feat_*.go *_feats_*.go
  python3 "$DOWNLOAD_SCRIPT" --min
)

# Step 3: Test that the minimal Wasmtime binary can deserialize and run a module.
go test -count=1 .
