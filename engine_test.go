package v2

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEngine(t *testing.T) {
	NewEngine()
	NewEngineWithConfig(NewConfig())
}

func TestEngineInvalidatesConfig(t *testing.T) {
	config := NewConfig()
	{
		engine := NewEngineWithConfig(config)
		require.NotNil(t, engine)
	}

	{
		defer func() {
			r := recover()
			require.NotNil(t, r, "The code did not panic")
		}()
		engine := NewEngineWithConfig(config)
		require.Nil(t, engine)
	}
}
