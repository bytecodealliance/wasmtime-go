package wasmtime

// #include <wasm.h>
// #include <wasmtime.h>
import "C"
import "errors"
import "runtime"

// Compilation strategies for wasmtime
type Strategy C.wasmtime_strategy_t

const (
	// Wasmtime will automatically pick an appropriate compilation strategy
	STRATEGY_AUTO = C.WASMTIME_STRATEGY_AUTO
	// Force wasmtime to use the Cranelift backend
	STRATEGY_CRANELIFT = C.WASMTIME_STRATEGY_CRANELIFT
	// Force wasmtime to use the lightbeam backend
	STRATEGY_LIGHTBEAM = C.WASMTIME_STRATEGY_LIGHTBEAM
)

// What degree of optimization wasmtime will perform on generated machine code
type OptLevel C.wasmtime_opt_level_t

const (
	// No optimizations will be performed
	OPT_LEVEL_NONE = C.WASMTIME_OPT_LEVEL_NONE
	// Machine code will be optimized to be as fast as possible
	OPT_LEVEL_SPEED = C.WASMTIME_OPT_LEVEL_SPEED
	// Machine code will be optimized for speed, but also optimized
	// to be small, sometimes at the cost of speed.
	OPT_LEVEL_SPEED_AND_SIZE = C.WASMTIME_OPT_LEVEL_SPEED_AND_SIZE
)

// What sort of profiling to enable, if any.
type ProfilingStrategy C.wasmtime_profiling_strategy_t

const (
	// No profiler will be used
	PROFILING_STRATEGY_NONE = C.WASMTIME_PROFILING_STRATEGY_NONE
        // The "jitdump" linux support will be used
	PROFILING_STRATEGY_JITDUMP = C.WASMTIME_PROFILING_STRATEGY_JITDUMP
)

// Configuration of an `Engine` which is used to globally configure things
// like wasm features and such.
type Config struct {
	_ptr *C.wasm_config_t
}

// Creates a new `Config` with all default options configured.
func NewConfig() *Config {
	config := &Config{_ptr: C.wasm_config_new()}
	runtime.SetFinalizer(config, func(config *Config) {
		C.wasm_config_delete(config._ptr)
	})
	return config
}

// Configures whether dwarf debug information for JIT code is enabled
func (cfg *Config) SetDebugInfo(enabled bool) {
	C.wasmtime_config_debug_info_set(cfg.ptr(), C.bool(enabled))
	runtime.KeepAlive(cfg)
}

// Configures whether the wasm threads proposal is enabled
func (cfg *Config) SetWasmThreads(enabled bool) {
	C.wasmtime_config_wasm_threads_set(cfg.ptr(), C.bool(enabled))
	runtime.KeepAlive(cfg)
}

// Configures whether the wasm reference types proposal is enabled
func (cfg *Config) SetWasmReferenceTypes(enabled bool) {
	C.wasmtime_config_wasm_reference_types_set(cfg.ptr(), C.bool(enabled))
	runtime.KeepAlive(cfg)
}

// Configures whether the wasm SIMD proposal is enabled
func (cfg *Config) SetWasmSIMD(enabled bool) {
	C.wasmtime_config_wasm_simd_set(cfg.ptr(), C.bool(enabled))
	runtime.KeepAlive(cfg)
}

// Configures whether the wasm bulk memory proposal is enabled
func (cfg *Config) SetWasmBulkMemory(enabled bool) {
	C.wasmtime_config_wasm_bulk_memory_set(cfg.ptr(), C.bool(enabled))
	runtime.KeepAlive(cfg)
}

// Configures whether the wasm multi value proposal is enabled
func (cfg *Config) SetWasmMultiValue(enabled bool) {
	C.wasmtime_config_wasm_multi_value_set(cfg.ptr(), C.bool(enabled))
	runtime.KeepAlive(cfg)
}

// Configures what compilation strategy is used to compile wasm code
func (cfg *Config) SetStrategy(strat Strategy) error {
	ok := C.wasmtime_config_strategy_set(cfg.ptr(), C.wasmtime_strategy_t(strat))
	runtime.KeepAlive(cfg)
	if !ok {
		return errors.New("failed to configure compilation strategy")
	}
	return nil
}

// Configures whether the cranelift debug verifier will be active when
// cranelift is used to compile wasm code.
func (cfg *Config) SetCraneliftDebugVerifier(enabled bool) {
	C.wasmtime_config_cranelift_debug_verifier_set(cfg.ptr(), C.bool(enabled))
	runtime.KeepAlive(cfg)
}

// Configures the cranelift optimization level for generated code
func (cfg *Config) SetCraneliftOptLevel(level OptLevel) {
	C.wasmtime_config_cranelift_opt_level_set(cfg.ptr(), C.wasmtime_opt_level_t(level))
	runtime.KeepAlive(cfg)
}

// Configures what profiler strategy to use for generated code
func (cfg *Config) SetProfiler(profiler ProfilingStrategy) error {
	ok := C.wasmtime_config_profiler_set(cfg.ptr(), C.wasmtime_profiling_strategy_t(profiler))
	runtime.KeepAlive(cfg)
	if !ok {
		return errors.New("failed to configure profiler strategy")
	}
	return nil
}

// See comments in `ffi.go` for what's going on here
func (config *Config) ptr() *C.wasm_config_t {
	ret := config._ptr
	maybeGC()
	if ret == nil {
		panic("Config has already been used up")
	}
	return ret
}
