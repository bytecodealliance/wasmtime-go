package wasmtime

// #include <stdlib.h>
// #include <wasm.h>
// #include <wasmtime.h>
import "C"
import "runtime"
import "unsafe"

type Trap struct {
	_ptr *C.wasm_trap_t
}

type Frame struct {
	_ptr   *C.wasm_frame_t
	_owner interface{}
}

// Creates a new `Trap` with the `name` and the type provided.
func NewTrap(store *Store, message string) *Trap {
	cs := C.CString(message)
	message_vec := C.wasm_byte_vec_t{
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

func (t *Trap) ptr() *C.wasm_trap_t {
	ret := t._ptr
	maybeGC()
	return ret
}

// Returns the name in the module this export type is exporting
func (t *Trap) Message() string {
	message := C.wasm_byte_vec_t{}
	C.wasm_trap_message(t.ptr(), &message)
	ret := C.GoStringN(message.data, C.int(message.size-1))
	runtime.KeepAlive(t)
	C.wasm_byte_vec_delete(&message)
	return ret
}

func (t *Trap) Error() string {
	return t.Message()
}

type frameList struct {
	vec C.wasm_frame_vec_t
}

func (t *Trap) Frames() []*Frame {
	frames := &frameList{}
	C.wasm_trap_trace(t.ptr(), &frames.vec)
	runtime.KeepAlive(t)
	runtime.SetFinalizer(frames, func(frames *frameList) {
		C.wasm_frame_vec_delete(&frames.vec)
	})

	ret := make([]*Frame, int(frames.vec.size))
	base := unsafe.Pointer(frames.vec.data)
	var ptr *C.wasm_frame_t
	for i := 0; i < int(frames.vec.size); i++ {
		ptr := *(**C.wasm_frame_t)(unsafe.Pointer(uintptr(base) + unsafe.Sizeof(ptr)*uintptr(i)))
		ret[i] = &Frame{
			_ptr:   ptr,
			_owner: frames,
		}
	}
	return ret
}

func (f *Frame) ptr() *C.wasm_frame_t {
	ret := f._ptr
	maybeGC()
	return ret
}

// Returns the function index in the wasm module that this frame represents
func (f *Frame) FuncIndex() uint32 {
	ret := C.wasm_frame_func_index(f.ptr())
	runtime.KeepAlive(f)
	return uint32(ret)
}

// Returns the name, if available, for this frame's function
func (f *Frame) FuncName() *string {
	ret := C.wasmtime_frame_func_name(f.ptr())
	if ret == nil {
		runtime.KeepAlive(f)
		return nil
	}
	str := C.GoStringN(ret.data, C.int(ret.size))
	runtime.KeepAlive(f)
	return &str
}

// Returns the name, if available, for this frame's module
func (f *Frame) ModuleName() *string {
	ret := C.wasmtime_frame_module_name(f.ptr())
	if ret == nil {
		runtime.KeepAlive(f)
		return nil
	}
	str := C.GoStringN(ret.data, C.int(ret.size))
	runtime.KeepAlive(f)
	return &str
}
