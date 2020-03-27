package wasmtime

import "testing"

func TestEngine(t *testing.T) {
	NewEngine()
	NewEngineWithConfig(NewConfig())
}
