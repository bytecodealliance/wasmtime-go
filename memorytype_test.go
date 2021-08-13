package wasmtime

import "testing"

func TestMemoryType(t *testing.T) {
	ty := NewMemoryType(0, true, 100)
	ty.Minimum()
	ty.Maximum()

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

func TestMemoryType64(t *testing.T) {
	ty := NewMemoryType64(0x100000000, true, 0x100000001)
	if ty.Minimum() != 0x100000000 {
		panic("bad limits")
	}
	present, max := ty.Maximum()
	if !present || max != 0x100000001 {
		panic("bad limits")
	}
}
