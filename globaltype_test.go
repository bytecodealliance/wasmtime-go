package wasmtime

import "testing"
import "runtime"

func TestGlobalType(t *testing.T) {
	ty := NewGlobalType(NewValType(KindI32), true)
	if ty.Content().Kind() != KindI32 {
		panic("invalid kind")
	}
	if !ty.Mutable() {
		panic("invalid mutable")
	}
	content := ty.Content()
	runtime.GC()
	if content.Kind() != KindI32 {
		panic("invalid kind")
	}
}
