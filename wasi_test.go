package wasmtime

import "testing"

func TestWasiConfig(t *testing.T) {
	config := NewWasiConfig()
	config.SetEnv([]string{"WASMTIME"}, []string{"GO"})
	store := NewStore(NewEngine())
	instance, err := NewWasiInstance(store, config, "wasi_snapshot_preview1")
	check(err)
	if instance == nil {
		panic("nil instance")
	}
}

func TestBadWasiName(t *testing.T) {
	config := NewWasiConfig()
	store := NewStore(NewEngine())
	instance, err := NewWasiInstance(store, config, "wrong_name")
	if err == nil {
		panic("expected an error")
	}
	if instance != nil {
		panic("not-nil instance")
	}
}
