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

	ty = NewGlobalType(NewValType(KindI32), true)
	ty2 := ty.AsExternType().GlobalType()
	if ty2 == nil {
		panic("unexpected cast")
	}
	if ty.AsExternType().FuncType() != nil {
		panic("working cast")
	}
	if ty.AsExternType().MemoryType() != nil {
		panic("working cast")
	}
	if ty.AsExternType().TableType() != nil {
		panic("working cast")
	}
}
