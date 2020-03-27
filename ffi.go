package wasmtime

// #cgo !windows LDFLAGS:-lwasmtime
// #cgo windows LDFLAGS:-lwasmtime.dll
import "C"
