package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEngine(t *testing.T) {
	NewEngine()
	NewEngineWithConfig(NewConfig())
}

func TestEngineInvalidatesConfig(t *testing.T) {
	config := NewConfig()
	{
		engine := NewEngineWithConfig(config)
		assert.NotNil(t, engine)
	}

	{
		defer func() {
			r := recover()
			assert.NotNil(t, r, "The code did not panic")
		}()
		engine := NewEngineWithConfig(config)
		assert.Nil(t, engine)
	}
}
