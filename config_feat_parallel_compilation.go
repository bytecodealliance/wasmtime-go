package wasmtime

// #include <wasmtime.h>
import "C"
import "runtime"

// SetParallelCompilation configures whether compilation should use multiple threads
func (cfg *Config) SetParallelCompilation(enabled bool) {
	C.wasmtime_config_parallel_compilation_set(cfg.ptr(), C.bool(enabled))
	runtime.KeepAlive(cfg)
}
