package wasmtime

import "testing"

func TestImportType(t *testing.T) {
	fty := NewFuncType([]*ValType{}, []*ValType{})
	ty := NewImportType("a", "b", fty)
	if ty.Module() != "a" {
		panic("invalid module")
	}
	if *ty.Name() != "b" {
		panic("invalid name")
	}
	if ty.Type().FuncType() == nil {
		panic("invalid ty")
	}

	gty := NewGlobalType(NewValType(KindI32), true)
	ty = NewImportType("", "", gty.AsExternType())
	if ty.Module() != "" {
		panic("invalid module")
	}
	if *ty.Name() != "" {
		panic("invalid name")
	}
	if ty.Type().GlobalType() == nil {
		panic("invalid ty")
	}
}
