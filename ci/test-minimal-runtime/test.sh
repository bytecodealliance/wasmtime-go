#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

WASMTIME_GO="$(cd "$SCRIPT_DIR/../.." && pwd)"
DOWNLOAD_SCRIPT="$WASMTIME_GO/ci/download-wasmtime.py"

# Step 1: Ensure the local wasmtime-go build directory has the v45 full libraries.
# The local build may contain libraries built from a different wasmtime version
# (e.g., a newer version built from local Rust source), which causes serialization format mismatches.
# We temporarily replace them with the official v45 release binaries.
trap 'rm -rf vendor module.cwasm' EXIT
(
  cd "$WASMTIME_GO"
  python3 "$DOWNLOAD_SCRIPT"
)

# Step 2: Create a pre-compiled module using the full Wasmtime library.
go run create_cwasm.go

# Step 3: Vendor the module, then adjust the vendored copy so it compiles
# against the minimal Wasmtime library:
#   a) Remove Go source files that call C functions absent from the min binary.
#   b) Download the min static libraries over the full ones so CGO links the right lib.
go mod vendor
(
  cd vendor/github.com/bytecodealliance/wasmtime-go/v45
  rm -f wat2wasm.go wasi.go *_feat_*.go *_feats_*.go
  python3 "$DOWNLOAD_SCRIPT" --min
)

# Step 4: Test that the minimal Wasmtime binary can deserialize and run a module.
go test -count=1 .
