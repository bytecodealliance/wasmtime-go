package wasmtime

// #include <wasi.h>
// #include <wasmtime.h>
// #include <stdlib.h>
// #include <stdint.h>
import "C"
import (
	"errors"
	"os"
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

// SetArgv will explicitly configure the argv for this WASI configuration.
// Note that this field can only be set, it cannot be read
func (c *WasiConfig) SetArgv(argv []string) {
	ptrs := make([]*C.char, len(argv))
	for i, arg := range argv {
		ptrs[i] = C.CString(arg)
	}
	var argvRaw **C.char
	if len(ptrs) > 0 {
		argvRaw = &ptrs[0]
	}
	C.wasi_config_set_argv(c.ptr(), C.int(len(argv)), argvRaw)
	runtime.KeepAlive(c)
	for _, ptr := range ptrs {
		C.free(unsafe.Pointer(ptr))
	}
}

func (c *WasiConfig) InheritArgv() {
	C.wasi_config_inherit_argv(c.ptr())
	runtime.KeepAlive(c)
}

// SetEnv configures environment variables to be returned for this WASI configuration.
// The pairs provided must be an iterable list of key/value pairs of environment variables.
// Note that this field can only be set, it cannot be read
func (c *WasiConfig) SetEnv(keys, values []string) {
	if len(keys) != len(values) {
		panic("mismatched numbers of keys and values")
	}
	namePtrs := make([]*C.char, len(values))
	valuePtrs := make([]*C.char, len(values))
	for i, key := range keys {
		namePtrs[i] = C.CString(key)
	}
	for i, value := range values {
		valuePtrs[i] = C.CString(value)
	}
	var namesRaw, valuesRaw **C.char
	if len(keys) > 0 {
		namesRaw = &namePtrs[0]
		valuesRaw = &valuePtrs[0]
	}
	C.wasi_config_set_env(c.ptr(), C.int(len(keys)), namesRaw, valuesRaw)
	runtime.KeepAlive(c)
	for i, ptr := range namePtrs {
		C.free(unsafe.Pointer(ptr))
		C.free(unsafe.Pointer(valuePtrs[i]))
	}
}

func (c *WasiConfig) InheritEnv() {
	C.wasi_config_inherit_env(c.ptr())
	runtime.KeepAlive(c)
}

func (c *WasiConfig) SetStdinFile(path string) error {
	pathC := C.CString(path)
	ok := C.wasi_config_set_stdin_file(c.ptr(), pathC)
	runtime.KeepAlive(c)
	C.free(unsafe.Pointer(pathC))
	if ok {
		return nil
	}

	return errors.New("failed to open file")
}

func (c *WasiConfig) InheritStdin() {
	C.wasi_config_inherit_stdin(c.ptr())
	runtime.KeepAlive(c)
}

func (c *WasiConfig) SetStdoutFile(path string) error {
	pathC := C.CString(path)
	ok := C.wasi_config_set_stdout_file(c.ptr(), pathC)
	runtime.KeepAlive(c)
	C.free(unsafe.Pointer(pathC))
	if ok {
		return nil
	}

	return errors.New("failed to open file")
}

func (c *WasiConfig) InheritStdout() {
	C.wasi_config_inherit_stdout(c.ptr())
	runtime.KeepAlive(c)
}

func (c *WasiConfig) SetStderrFile(path string) error {
	pathC := C.CString(path)
	ok := C.wasi_config_set_stderr_file(c.ptr(), pathC)
	runtime.KeepAlive(c)
	C.free(unsafe.Pointer(pathC))
	if ok {
		return nil
	}

	return errors.New("failed to open file")
}

func (c *WasiConfig) InheritStderr() {
	C.wasi_config_inherit_stderr(c.ptr())
	runtime.KeepAlive(c)
}

func (c *WasiConfig) PreopenDir(path, guestPath string) error {
	pathC := C.CString(path)
	guestPathC := C.CString(guestPath)
	ok := C.wasi_config_preopen_dir(c.ptr(), pathC, guestPathC)
	runtime.KeepAlive(c)
	C.free(unsafe.Pointer(pathC))
	C.free(unsafe.Pointer(guestPathC))
	if ok {
		return nil
	}

	return errors.New("failed to preopen directory")
}

// PreopenTCPSocket configures a "preopened" tcp listen socket to be available to
// WASI APIs after build.
//
// By default WASI programs do not have access to open up network sockets on the
// host. This API can be used to grant WASI programs access to a network socket
// file descriptor on the host.
//
// The fd_num argument is the number of the file descriptor by which it will be
// known in WASM and the host_port is the IP address and port (e.g.
// "127.0.0.1:8080") requested to listen on.
func (c *WasiConfig) PreopenTCPSocket(innerFD uint32, hostPort string) error {

	c_hostPort := C.CString(hostPort)

	ok := C.wasi_config_preopen_socket(c.ptr(), C.uint32_t(innerFD), c_hostPort)
	runtime.KeepAlive(c)
	C.free(unsafe.Pointer(&hostPort))
	if ok {
		return nil
	}
	return errors.New("failed to conifgure preopen tcp socket")
}

// FileAccessMode Indicates whether the file-like object being inserted into the
// WASI configuration (by PushFile and InsertFile) can be used to read, write,
// or both using bitflags. This seems to be a wasmtime specific mapping as it
// does not match syscall.O_RDONLY, O_WRONLY, etc.
type WasiFileAccessMode uint32

const (
	READ WasiFileAccessMode = 1 << iota
	WRITE
	READ_WRITE = READ | WRITE
)

// WasiCtx maps to the Rust `WasiCtx` type.
//
// Instead of wrapping a pointer to the type directly, this type is a wrapper
// around a Store in order to call the `wasmtime_context_*` functions.
type WasiCtx struct {
	store *Store
}

func (ctx *WasiCtx) InsertFile(guestFD uint32, file *os.File, accessMode WasiFileAccessMode) {
	C.wasmtime_context_insert_file(ctx.store.Context(), C.uint32_t(guestFD), unsafe.Pointer(file.Fd()), C.uint32_t(accessMode))
	runtime.KeepAlive(ctx.store)
	runtime.KeepAlive(file)
}

func (ctx *WasiCtx) PushFile(file *os.File, accessMode WasiFileAccessMode) (uint32, error) {
	var fd uint32
	c_fd := C.uint32_t(fd)

	err := C.wasmtime_context_push_file(ctx.store.Context(), unsafe.Pointer(file.Fd()), C.uint32_t(accessMode), &c_fd)
	runtime.KeepAlive(ctx.store)
	runtime.KeepAlive(file)
	if err != nil {
		return 0, mkError(err)
	}

	return uint32(c_fd), nil
}
