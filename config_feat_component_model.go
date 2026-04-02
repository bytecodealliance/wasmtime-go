package wasmtime

// #include <wasmtime.h>
import "C"
import "runtime"

// SetWasmComponentModel configures whether the wasm component model proposal is
// enabled.
func (cfg *Config) SetWasmComponentModel(enabled bool) {
	C.wasmtime_config_wasm_component_model_set(cfg.ptr(), C.bool(enabled))
	runtime.KeepAlive(cfg)
}
