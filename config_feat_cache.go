package wasmtime

// #include <wasmtime.h>
// #include <stdlib.h>
import "C"
import (
	"runtime"
	"unsafe"
)

// CacheConfigLoadDefault enables compiled code caching for this `Config` using the default settings
// configuration can be found.
//
// For more information about caching see
// https://bytecodealliance.github.io/wasmtime/cli-cache.html
func (cfg *Config) CacheConfigLoadDefault() error {
	err := C.wasmtime_config_cache_config_load(cfg.ptr(), nil)
	runtime.KeepAlive(cfg)
	if err != nil {
		return mkError(err)
	}
	return nil
}

// CacheConfigLoad enables compiled code caching for this `Config` using the settings specified
// in the configuration file `path`.
//
// For more information about caching and configuration options see
// https://bytecodealliance.github.io/wasmtime/cli-cache.html
func (cfg *Config) CacheConfigLoad(path string) error {
	cstr := C.CString(path)
	err := C.wasmtime_config_cache_config_load(cfg.ptr(), cstr)
	C.free(unsafe.Pointer(cstr))
	runtime.KeepAlive(cfg)
	if err != nil {
		return mkError(err)
	}
	return nil
}
