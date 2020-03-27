package wasmtime

// #include <wasm.h>
import "C"
import "runtime"

type Engine struct {
	_ptr *C.wasm_engine_t
}

func NewEngine() *Engine {
	engine := &Engine{_ptr: C.wasm_engine_new()}
	runtime.SetFinalizer(engine, func(engine *Engine) {
		C.wasm_engine_delete(engine._ptr)
	})
	return engine
}

func NewEngineWithConfig(config *Config) *Engine {
	if config.ptr() == nil {
		panic("config already used")
	}
	engine := &Engine{_ptr: C.wasm_engine_new_with_config(config.ptr())}
	runtime.SetFinalizer(config, nil)
	runtime.SetFinalizer(engine, func(engine *Engine) {
		C.wasm_engine_delete(engine._ptr)
	})
	return engine
}

func (engine *Engine) ptr() *C.wasm_engine_t {
	ret := engine._ptr
	maybeGC()
	return ret
}
