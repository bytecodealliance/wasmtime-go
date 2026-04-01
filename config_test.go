package wasmtime

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	NewConfig().SetDebugInfo(true)
	NewConfig().SetMaxWasmStack(8388608)
	NewConfig().SetWasmThreads(true)
	NewConfig().SetWasmReferenceTypes(true)
	NewConfig().SetWasmSIMD(true)
	NewConfig().SetWasmRelaxedSIMD(true)
	NewConfig().SetWasmRelaxedSIMDDeterministic(true)
	NewConfig().SetWasmBulkMemory(true)
	NewConfig().SetWasmMultiValue(true)
	NewConfig().SetWasmMultiMemory(true)
	NewConfig().SetWasmTailCall(true)
	NewConfig().SetWasmFunctionReferences(true)
	NewConfig().SetWasmGC(true)
	NewConfig().SetGCSupport(true)
	NewConfig().SetWasmComponentModel(true)
	NewConfig().SetWasmWideArithmetic(true)
	NewConfig().SetConsumeFuel(true)
	NewConfig().SetStrategy(StrategyAuto)
	NewConfig().SetStrategy(StrategyCranelift)
	NewConfig().SetCraneliftDebugVerifier(true)
	NewConfig().SetCraneliftOptLevel(OptLevelNone)
	NewConfig().SetCraneliftOptLevel(OptLevelSpeed)
	NewConfig().SetCraneliftOptLevel(OptLevelSpeedAndSize)
	NewConfig().SetProfiler(ProfilingStrategyNone)
	NewConfig().SetParallelCompilation(true)
	NewConfig().SetCraneliftNanCanonicalization(true)
	NewConfig().SetNativeUnwindInfo(true)
	NewConfig().SetMacOSUseMachPorts(false)
	NewConfig().SetMemoryInitCOWSet(true)
	if runtime.GOARCH == "amd64" && runtime.GOOS == "linux" {
		NewConfig().SetTarget("x86_64-unknown-linux-gnu")
	}
	NewConfig().SetCraneliftFlag("opt_level", "none")
	NewConfig().EnableCraneliftFlag("unwind_info")
	err := NewConfig().CacheConfigLoadDefault()
	require.NoError(t, err)
	err = NewConfig().CacheConfigLoad("nonexistent.toml")
	require.Error(t, err)
}
