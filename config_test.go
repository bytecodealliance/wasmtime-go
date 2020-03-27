package wasmtime

import "testing"

func TestConfig(t *testing.T) {
	NewConfig().SetDebugInfo(true)
	NewConfig().SetWasmThreads(true)
	NewConfig().SetWasmReferenceTypes(true)
	NewConfig().SetWasmSIMD(true)
	NewConfig().SetWasmBulkMemory(true)
	NewConfig().SetWasmMultiValue(true)
	err := NewConfig().SetStrategy(STRATEGY_AUTO)
	if err != nil {
		panic(err)
	}
	err = NewConfig().SetStrategy(STRATEGY_CRANELIFT)
	if err != nil {
		panic(err)
	}
	NewConfig().SetCraneliftDebugVerifier(true)
	NewConfig().SetCraneliftOptLevel(OPT_LEVEL_NONE)
	NewConfig().SetCraneliftOptLevel(OPT_LEVEL_SPEED)
	NewConfig().SetCraneliftOptLevel(OPT_LEVEL_SPEED_AND_SIZE)
	NewConfig().SetProfiler(PROFILING_STRATEGY_NONE)
}
