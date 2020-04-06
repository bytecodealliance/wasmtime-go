package wasmtime

import "testing"

func TestEngine(t *testing.T) {
	NewEngine()
	NewEngineWithConfig(NewConfig())
}

func TestEngineInvalidatesConfig(t *testing.T) {
	config := NewConfig()
	NewEngineWithConfig(config)
	(func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		NewEngineWithConfig(config)
	})()
}
