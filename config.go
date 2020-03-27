package wasmtime

// #include <wasm.h>
import "C"
import "runtime"

type Config struct {
	_ptr *C.wasm_config_t
}

func NewConfig() *Config {
	config := &Config{_ptr: C.wasm_config_new()}
	runtime.SetFinalizer(config, func(config *Config) {
		C.wasm_config_delete(config._ptr)
	})
	return config
}

// See comments in `ffi.go` for what's going on here
func (config *Config) ptr() *C.wasm_config_t {
	ret := config._ptr
	maybeGC()
	return ret
}
