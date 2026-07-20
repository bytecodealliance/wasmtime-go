package wasmtime

// #include <wasm.h>
// #include <wasmtime.h>
// #include <stdlib.h>
import "C"
import "runtime"

// SetWasmGC configures whether garbage collection is enabled
func (cfg *Config) SetWasmGC(enabled bool) {
	C.wasmtime_config_wasm_gc_set(cfg.ptr(), C.bool(enabled))
	runtime.KeepAlive(cfg)
}

// SetWasmReferenceTypes configures whether the wasm reference types proposal is enabled
func (cfg *Config) SetWasmReferenceTypes(enabled bool) {
	C.wasmtime_config_wasm_reference_types_set(cfg.ptr(), C.bool(enabled))
	runtime.KeepAlive(cfg)
}

// SetWasmFunctionReferences configures whether function references are enabled
func (cfg *Config) SetWasmFunctionReferences(enabled bool) {
	C.wasmtime_config_wasm_function_references_set(cfg.ptr(), C.bool(enabled))
	runtime.KeepAlive(cfg)
}

// SetWasmExceptions configures whether exceptions are enabled
func (cfg *Config) SetWasmExceptions(enabled bool) {
	C.wasmtime_config_wasm_exceptions_set(cfg.ptr(), C.bool(enabled))
	runtime.KeepAlive(cfg)
}
