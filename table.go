package wasmtime

// #include <wasm.h>
import "C"
import "runtime"

type Table struct {
	_ptr   *C.wasm_table_t
	_owner interface{}
}

func mkTable(ptr *C.wasm_table_t, owner interface{}) *Table {
	f := &Table{_ptr: ptr, _owner: owner}
	if owner == nil {
		runtime.SetFinalizer(f, func(f *Table) {
			C.wasm_table_delete(f._ptr)
		})
	}
	return f
}

func (t *Table) ptr() *C.wasm_table_t {
	ret := t._ptr
	maybeGC()
	return ret
}

func (t *Table) owner() interface{} {
	if t._owner != nil {
		return t._owner
	}
	return t
}

func (t *Table) Size() uint32 {
	ret := C.wasm_table_size(t.ptr())
	runtime.KeepAlive(t)
	return uint32(ret)
}

func (t *Table) Type() *TableType {
	ptr := C.wasm_table_type(t.ptr())
	runtime.KeepAlive(t)
	return mkTableType(ptr, nil)
}

func (t *Table) AsExtern() *Extern {
	ptr := C.wasm_table_as_extern(t.ptr())
	return mkExtern(ptr, t.owner())
}
