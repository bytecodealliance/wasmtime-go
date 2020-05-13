package wasmtime

// #include <wasi.h>
// #include <stdlib.h>
import "C"
import (
	"errors"
	"runtime"
	"unsafe"
)

type WasiConfig struct {
	_ptr *C.wasi_config_t
}

func NewWasiConfig() *WasiConfig {
	ptr := C.wasi_config_new()
	config := &WasiConfig{_ptr: ptr}
	runtime.SetFinalizer(config, func(config *WasiConfig) {
		C.wasi_config_delete(config._ptr)
	})
	return config
}

func (c *WasiConfig) ptr() *C.wasi_config_t {
	ret := c._ptr
	maybeGC()
	return ret
}

func (c *WasiConfig) SetArgv(argv []string) {
	ptrs := make([]*C.char, len(argv))
	for i, arg := range argv {
		ptrs[i] = C.CString(arg)
	}
	var argv_raw **C.char
	if len(ptrs) > 0 {
		argv_raw = &ptrs[0]
	}
	C.wasi_config_set_argv(c.ptr(), C.int(len(argv)), argv_raw)
	runtime.KeepAlive(c)
	for _, ptr := range ptrs {
		C.free(unsafe.Pointer(ptr))
	}
}

func (c *WasiConfig) InheritArgv() {
	C.wasi_config_inherit_argv(c.ptr())
	runtime.KeepAlive(c)
}

func (c *WasiConfig) SetEnv(keys, values []string) {
	if len(keys) != len(values) {
		panic("mismatched numbers of keys and values")
	}
	name_ptrs := make([]*C.char, len(values))
	value_ptrs := make([]*C.char, len(values))
	for i, key := range keys {
		name_ptrs[i] = C.CString(key)
	}
	for i, value := range values {
		value_ptrs[i] = C.CString(value)
	}
	var names_raw, values_raw **C.char
	if len(keys) > 0 {
		names_raw = &name_ptrs[0]
		values_raw = &value_ptrs[0]
	}
	C.wasi_config_set_env(c.ptr(), C.int(len(keys)), names_raw, values_raw)
	runtime.KeepAlive(c)
	for i, ptr := range name_ptrs {
		C.free(unsafe.Pointer(ptr))
		C.free(unsafe.Pointer(value_ptrs[i]))
	}
}

func (c *WasiConfig) InheritEnv() {
	C.wasi_config_inherit_env(c.ptr())
	runtime.KeepAlive(c)
}

func (c *WasiConfig) SetStdinFile(path string) error {
	path_c := C.CString(path)
	ok := C.wasi_config_set_stdin_file(c.ptr(), path_c)
	runtime.KeepAlive(c)
	C.free(unsafe.Pointer(path_c))
	if ok {
		return nil
	} else {
		return errors.New("failed to open file")
	}
}

func (c *WasiConfig) InheritStdin() {
	C.wasi_config_inherit_stdin(c.ptr())
	runtime.KeepAlive(c)
}

func (c *WasiConfig) SetStdoutFile(path string) error {
	path_c := C.CString(path)
	ok := C.wasi_config_set_stdout_file(c.ptr(), path_c)
	runtime.KeepAlive(c)
	C.free(unsafe.Pointer(path_c))
	if ok {
		return nil
	} else {
		return errors.New("failed to open file")
	}
}

func (c *WasiConfig) InheritStdout() {
	C.wasi_config_inherit_stdout(c.ptr())
	runtime.KeepAlive(c)
}

func (c *WasiConfig) SetStderrFile(path string) error {
	path_c := C.CString(path)
	ok := C.wasi_config_set_stderr_file(c.ptr(), path_c)
	runtime.KeepAlive(c)
	C.free(unsafe.Pointer(path_c))
	if ok {
		return nil
	} else {
		return errors.New("failed to open file")
	}
}

func (c *WasiConfig) InheritStderr() {
	C.wasi_config_inherit_stderr(c.ptr())
	runtime.KeepAlive(c)
}

func (c *WasiConfig) PreopenDir(path, guest_path string) error {
	path_c := C.CString(path)
	guest_path_c := C.CString(guest_path)
	ok := C.wasi_config_preopen_dir(c.ptr(), path_c, guest_path_c)
	runtime.KeepAlive(c)
	C.free(unsafe.Pointer(path_c))
	C.free(unsafe.Pointer(guest_path_c))
	if ok {
		return nil
	} else {
		return errors.New("failed to preopen directory")
	}
}

type WasiInstance struct {
	_ptr     *C.wasi_instance_t
	freelist *freeList
}

// Creates a new instance of WASI with the given configuration.
//
// The version of WASI must be explicitly requested via `name`.
func NewWasiInstance(store *Store, config *WasiConfig, name string) (*WasiInstance, error) {
	if config._ptr == nil {
		panic("config already used to create wasi instance")
	}
	var trap *C.wasm_trap_t
	name_ptr := C.CString(name)
	ptr := C.wasi_instance_new(
		store.ptr(),
		name_ptr,
		config.ptr(),
		&trap,
	)
	runtime.KeepAlive(store)
	config._ptr = nil
	runtime.SetFinalizer(config, nil)
	C.free(unsafe.Pointer(name_ptr))

	if ptr == nil {
		if trap != nil {
			return nil, mkTrap(trap)
		}
		return nil, errors.New("failed to create instance")
	}

	instance := &WasiInstance{
		_ptr:     ptr,
		freelist: store.freelist,
	}
	runtime.SetFinalizer(instance, func(instance *WasiInstance) {
		freelist := instance.freelist
		freelist.lock.Lock()
		defer freelist.lock.Unlock()
		freelist.wasi_instances = append(freelist.wasi_instances, instance._ptr)
	})
	return instance, nil
}

func (i *WasiInstance) ptr() *C.wasi_instance_t {
	ret := i._ptr
	maybeGC()
	return ret
}

// Attempts to bind the `imp` import provided, returning an Extern suitable for
// satisfying the import if one can be found.
//
// If `imp` isn't defined by this instance of WASI then `nil` is returned.
func (i *WasiInstance) BindImport(imp *ImportType) *Extern {
	ret := C.wasi_instance_bind_import(i.ptr(), imp.ptr())
	runtime.KeepAlive(i)
	runtime.KeepAlive(imp)
	if ret == nil {
		return nil
	} else {
		return mkExtern(ret, i.freelist, nil)
	}
}
