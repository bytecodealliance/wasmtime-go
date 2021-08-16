package wasmtime

import "testing"

func TestTableType(t *testing.T) {
	ty := NewTableType(NewValType(KindI32), 0, false, 0)
	if ty.Element().Kind() != KindI32 {
		panic("invalid kind")
	}
	if ty.Minimum() != 0 {
		panic("invalid min")
	}
	present, _ := ty.Maximum()
	if present {
		panic("invalid max")
	}

	ty = NewTableType(NewValType(KindF64), 1, true, 129)
	if ty.Element().Kind() != KindF64 {
		panic("invalid kind")
	}
	if ty.Minimum() != 1 {
		panic("invalid min")
	}
	present, max := ty.Maximum()
	if !present || max != 129 {
		panic("invalid max")
	}

	ty2 := ty.AsExternType().TableType()
	if ty2 == nil {
		panic("unexpected cast")
	}
	if ty.AsExternType().FuncType() != nil {
		panic("working cast")
	}
	if ty.AsExternType().GlobalType() != nil {
		panic("working cast")
	}
	if ty.AsExternType().MemoryType() != nil {
		panic("working cast")
	}
}
