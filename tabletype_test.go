package wasmtime

import "testing"

func TestTableType(t *testing.T) {
	ty := NewTableType(NewValType(KindI32), Limits{})
	if ty.Element().Kind() != KindI32 {
		panic("invalid kind")
	}
	if ty.Limits().Min != 0 {
		panic("invalid min")
	}
	if ty.Limits().Max != 0 {
		panic("invalid max")
	}

	ty = NewTableType(NewValType(KindF64), Limits{Min: 1, Max: 129})
	if ty.Element().Kind() != KindF64 {
		panic("invalid kind")
	}
	if ty.Limits().Min != 1 {
		panic("invalid min")
	}
	if ty.Limits().Max != 129 {
		panic("invalid max")
	}
}
