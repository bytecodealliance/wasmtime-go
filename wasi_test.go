package wasmtime

import "testing"

func TestWasiConfig(t *testing.T) {
	config := NewWasiConfig()
	config.SetEnv([]string{"WASMTIME"}, []string{"GO"})
}
