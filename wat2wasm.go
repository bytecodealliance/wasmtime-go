package wasmtime

// #include <wasmtime.h>
import "C"
import "errors"
import "runtime"
import "unsafe"

func Wat2Wasm(wat string) ([]byte, error) {
	wat_vec := C.wasm_byte_vec_t{
		size: C._GoStringLen(wat),
		data: C._GoStringPtr(wat),
	}

	ret_vec := C.wasm_byte_vec_t{}
	ret_error := C.wasm_byte_vec_t{}
	ok := C.wasmtime_wat2wasm(&wat_vec, &ret_vec, &ret_error)
	runtime.KeepAlive(wat)

	if ok {
		ret := C.GoBytes(unsafe.Pointer(ret_vec.data), C.int(ret_vec.size))
		C.wasm_byte_vec_delete(&ret_vec)
		return ret, nil
	} else {
		message := C.GoStringN(ret_error.data, C.int(ret_error.size))
		C.wasm_byte_vec_delete(&ret_error)
		return nil, errors.New(message)
	}
}
