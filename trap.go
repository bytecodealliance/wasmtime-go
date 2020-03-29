package wasmtime

// #include <stdlib.h>
// #include <wasm.h>
import "C"
import "runtime"
import "unsafe"

type Trap struct {
	_ptr *C.wasm_trap_t
}

// Creates a new `Trap` with the `name` and the type provided.
func NewTrap(store *Store, message string) *Trap {
        cs := C.CString(message)
	message_vec := C.wasm_byte_vec_t {
          data: cs,
          size: C.size_t(len(message) + 1),
        }
	ptr := C.wasm_trap_new(store.ptr(), &message_vec)
        C.free(unsafe.Pointer(cs))
	runtime.KeepAlive(store)
	return mkTrap(ptr)
}

func mkTrap(ptr *C.wasm_trap_t) *Trap {
	trap := &Trap{_ptr: ptr}
	runtime.SetFinalizer(trap, func(trap *Trap) {
		C.wasm_trap_delete(trap._ptr)
	})
	return trap
}

func (ty *Trap) ptr() *C.wasm_trap_t {
	ret := ty._ptr
	maybeGC()
	return ret
}

// Returns the name in the module this export type is exporting
func (ty *Trap) Message() string {
	message := C.wasm_byte_vec_t{}
	C.wasm_trap_message(ty.ptr(), &message)
	ret := C.GoStringN(message.data, C.int(message.size - 1))
	runtime.KeepAlive(ty)
	C.wasm_byte_vec_delete(&message)
	return ret
}

func (ty *Trap) Error() string {
        return ty.Message()
}
