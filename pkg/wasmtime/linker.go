package wasmtime

// #include <wasmtime.h>
import "C"
import "runtime"
import "errors"

type Linker struct {
	_ptr  *C.wasmtime_linker_t
	Store *Store
}

func NewLinker(store *Store) *Linker {
	ptr := C.wasmtime_linker_new(store.ptr())
	runtime.KeepAlive(store)
	return mkLinker(ptr, store)
}

func mkLinker(ptr *C.wasmtime_linker_t, store *Store) *Linker {
	linker := &Linker{_ptr: ptr, Store: store}
	runtime.SetFinalizer(linker, func(linker *Linker) {
		C.wasmtime_linker_delete(linker._ptr)
	})
	return linker
}

func (l *Linker) ptr() *C.wasmtime_linker_t {
	ret := l._ptr
	maybeGC()
	return ret
}

func (l *Linker) AllowShadowing(allow bool) {
	C.wasmtime_linker_allow_shadowing(l.ptr(), C.bool(allow))
	runtime.KeepAlive(l)
}

func (l *Linker) Define(module, name string, item AsExtern) bool {
	extern := item.AsExtern()
	module_vec := stringToBorrowedByteVec(module)
	name_vec := stringToBorrowedByteVec(name)
	ret := C.wasmtime_linker_define(l.ptr(), &module_vec, &name_vec, extern.ptr())
	runtime.KeepAlive(l)
	runtime.KeepAlive(module)
	runtime.KeepAlive(name)
	runtime.KeepAlive(extern)
	return bool(ret)
}

func (l *Linker) DefineFunc(module, name string, f interface{}) bool {
	return l.Define(module, name, WrapFunc(l.Store, f))
}

func (l *Linker) DefineInstance(name string, instance *Instance) bool {
	name_vec := stringToBorrowedByteVec(name)
	ret := C.wasmtime_linker_define_instance(l.ptr(), &name_vec, instance.ptr())
	runtime.KeepAlive(l)
	runtime.KeepAlive(name)
	runtime.KeepAlive(instance)
	return bool(ret)
}

// func (l *Linker) DefineWasi(instance *WasiInstance) bool {
// 	ret := C.wasmtime_linker_define_wasi(l.ptr(), instance.ptr())
// 	runtime.KeepAlive(l)
// 	runtime.KeepAlive(instance)
// 	return bool(ret)
// }

func (l *Linker) Instantiate(module *Module) (*Instance, error) {
	var trap *C.wasm_trap_t
	ret := C.wasmtime_linker_instantiate(l.ptr(), module.ptr(), &trap)
	runtime.KeepAlive(l)
	runtime.KeepAlive(module)
	if ret == nil {
		if trap != nil {
			return nil, mkTrap(trap)
		}
		return nil, errors.New("failed to instantiate")
	}
	return mkInstance(ret), nil
}
