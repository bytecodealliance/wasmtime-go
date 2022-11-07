package v2

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	NewConfig().SetDebugInfo(true)
	NewConfig().SetWasmThreads(true)
	NewConfig().SetWasmReferenceTypes(true)
	NewConfig().SetWasmSIMD(true)
	NewConfig().SetWasmBulkMemory(true)
	NewConfig().SetWasmMultiValue(true)
	NewConfig().SetWasmMultiMemory(true)
	NewConfig().SetConsumeFuel(true)
	NewConfig().SetStrategy(StrategyAuto)
	NewConfig().SetStrategy(StrategyCranelift)
	NewConfig().SetCraneliftDebugVerifier(true)
	NewConfig().SetCraneliftOptLevel(OptLevelNone)
	NewConfig().SetCraneliftOptLevel(OptLevelSpeed)
	NewConfig().SetCraneliftOptLevel(OptLevelSpeedAndSize)
	NewConfig().SetProfiler(ProfilingStrategyNone)
	err := NewConfig().CacheConfigLoadDefault()
	require.NoError(t, err)
	err = NewConfig().CacheConfigLoad("nonexistent.toml")
	require.Error(t, err)
}
