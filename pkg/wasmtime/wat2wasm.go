package wasmtime

// #include <wasmtime.h>
//
// bool go_wat2wasm(
//  char *wat_ptr,
//  size_t wat_len,
//  wasm_byte_vec_t *ret,
//  wasm_byte_vec_t *error_message
// ) {
//   wasm_byte_vec_t wat;
//   wat.data = wat_ptr;
//   wat.size = wat_len;
//   return wasmtime_wat2wasm(&wat, ret, error_message);
// }
import "C"
import "errors"
import "runtime"
import "unsafe"

func Wat2Wasm(wat string) ([]byte, error) {
	ret_vec := C.wasm_byte_vec_t{}
	ret_error := C.wasm_byte_vec_t{}
	ok := C.go_wat2wasm(
		C._GoStringPtr(wat),
		C._GoStringLen(wat),
		&ret_vec,
		&ret_error,
	)
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
