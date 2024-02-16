package wasmtime

import "testing"

func TestWasiConfig(t *testing.T) {
	config := NewWasiConfig()
	defer config.Close()
	config.SetEnv([]string{"WASMTIME"}, []string{"GO"})
}
