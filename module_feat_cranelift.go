package wasmtime

// #include "shims.h"
import "C"

import (
	"runtime"
	"unsafe"
)

// NewModule compiles a new `Module` from the `wasm` provided with the given configuration
// in `engine`.
func NewModule(engine *Engine, wasm []byte) (*Module, error) {
	// We can't create the `wasm_byte_vec_t` here and pass it in because
	// that runs into the error of "passed a pointer to a pointer" because
	// the vec itself is passed by pointer and it contains a pointer to
	// `wasm`. To work around this we insert some C shims above and call
	// them.
	var wasmPtr *C.uint8_t
	if len(wasm) > 0 {
		wasmPtr = (*C.uint8_t)(unsafe.Pointer(&wasm[0]))
	}
	var ptr *C.wasmtime_module_t
	err := C.wasmtime_module_new(engine.ptr(), wasmPtr, C.size_t(len(wasm)), &ptr)
	runtime.KeepAlive(engine)
	runtime.KeepAlive(wasm)

	if err != nil {
		return nil, mkError(err)
	}

	return mkModule(ptr), nil
}

// ModuleValidate validates whether `wasm` would be a valid wasm module according to the
// configuration in `store`
func ModuleValidate(engine *Engine, wasm []byte) error {
	var wasmPtr *C.uint8_t
	if len(wasm) > 0 {
		wasmPtr = (*C.uint8_t)(unsafe.Pointer(&wasm[0]))
	}
	err := C.wasmtime_module_validate(engine.ptr(), wasmPtr, C.size_t(len(wasm)))
	runtime.KeepAlive(engine)
	runtime.KeepAlive(wasm)
	if err == nil {
		return nil
	}

	return mkError(err)
}

// Serialize will convert this in-memory compiled module into a list of bytes.
//
// The purpose of this method is to extract an artifact which can be stored
// elsewhere from this `Module`. The returned bytes can, for example, be stored
// on disk or in an object store. The `NewModuleDeserialize` function can be
// used to deserialize the returned bytes at a later date to get the module
// back.
func (m *Module) Serialize() ([]byte, error) {
	retVec := C.wasm_byte_vec_t{}
	err := C.wasmtime_module_serialize(m.ptr(), &retVec)
	runtime.KeepAlive(m)

	if err != nil {
		return nil, mkError(err)
	}
	ret := C.GoBytes(unsafe.Pointer(retVec.data), C.int(retVec.size))
	C.wasm_byte_vec_delete(&retVec)
	return ret, nil
}
