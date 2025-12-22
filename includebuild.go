//go:build includebuild
// +build includebuild

package wasmtime

// This file is not built and not included in BUILD.bazel;
// it is only used to prevent "go mod vendor" to prune the
// build directory.

import (
	// Import these build directories in order to have them
	// included in vendored dependencies.
	// Cf. https://github.com/golang/go/issues/26366

	_ "github.com/bytecodealliance/wasmtime-go/v40/build/include"
	_ "github.com/bytecodealliance/wasmtime-go/v40/build/include/wasmtime"
	_ "github.com/bytecodealliance/wasmtime-go/v40/build/include/wasmtime/component"
	_ "github.com/bytecodealliance/wasmtime-go/v40/build/include/wasmtime/component/types"
	_ "github.com/bytecodealliance/wasmtime-go/v40/build/include/wasmtime/types"
	_ "github.com/bytecodealliance/wasmtime-go/v40/build/linux-aarch64"
	_ "github.com/bytecodealliance/wasmtime-go/v40/build/linux-x86_64"
	_ "github.com/bytecodealliance/wasmtime-go/v40/build/macos-aarch64"
	_ "github.com/bytecodealliance/wasmtime-go/v40/build/macos-x86_64"
	_ "github.com/bytecodealliance/wasmtime-go/v40/build/windows-x86_64"
)
