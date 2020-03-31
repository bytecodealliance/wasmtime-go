package wasmtime

import "testing"

func TestExportType(t *testing.T) {
	et := NewExportType("x", NewMemoryType(Limits{}))
	if et.Name() != "x" {
		panic("bad name")
	}
	if et.Type().MemoryType() == nil {
		panic("bad type")
	}
}
