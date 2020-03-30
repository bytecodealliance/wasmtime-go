package wasmtime

import "testing"

func TestMemoryType(t *testing.T) {
	ty := NewMemoryType(Limits{Min: 0, Max: 100})
	ty.Limits()

	ty2 := ty.AsExtern().MemoryType()
	if ty2 == nil {
		panic("unexpected cast")
	}
	if ty.AsExtern().FuncType() != nil {
		panic("working cast")
	}
	if ty.AsExtern().GlobalType() != nil {
		panic("working cast")
	}
	if ty.AsExtern().TableType() != nil {
		panic("working cast")
	}
}
