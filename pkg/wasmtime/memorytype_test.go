package wasmtime

import "testing"

func TestMemoryType(t *testing.T) {
	ty := NewMemoryType(Limits{Min: 0, Max: 100})
	ty.Limits()

	ty2 := ty.AsExternType().MemoryType()
	if ty2 == nil {
		panic("unexpected cast")
	}
	if ty.AsExternType().FuncType() != nil {
		panic("working cast")
	}
	if ty.AsExternType().GlobalType() != nil {
		panic("working cast")
	}
	if ty.AsExternType().TableType() != nil {
		panic("working cast")
	}
}
