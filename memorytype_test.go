package wasmtime

import "testing"

func TestMemoryType(t *testing.T) {
	ty := NewMemoryType(Limits{Min: 0, Max: 100})
	ty.Limits()
}
