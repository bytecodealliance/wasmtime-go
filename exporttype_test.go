package wasmtime

import "testing"

func TestExportType(t *testing.T) {
	et := NewExportType("x", NewMemoryType(0, false, 0))
	if et.Name() != "x" {
		panic("bad name")
	}
	if et.Type().MemoryType() == nil {
		panic("bad type")
	}
}
