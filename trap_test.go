package wasmtime

import "testing"

func TestTrap(t *testing.T) {
	trap := NewTrap(NewStore(NewEngine()), "message")
	if trap.Message() != "message" {
		panic("wrong message")
	}
}
