// +build debug

package wasmtime

import "runtime"

func maybeGC() {
	runtime.GC()
}
