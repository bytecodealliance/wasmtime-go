package wasmtime

// #include <wasmtime.h>
import "C"
import (
	"errors"
	"runtime"
)

// Table is a table instance, which is the runtime representation of a table.
// It holds a vector of function elements and an optional maximum size, if one was specified in the table type at the table’s definition site.
// Read more in [spec](https://webassembly.github.io/spec/core/exec/runtime.html#table-instances)
type Table struct {
	_ptr     *C.wasm_table_t
	_owner   interface{}
	freelist *freeList
}

// NewTable creates a new `Table` in the given `Store` with the specified `ty`.
//
// The `ty` must be a `funcref` table and `init` is the initial value for all
// table slots, and is allowed to be `nil`.
func NewTable(store *Store, ty *TableType, init *Func) (*Table, error) {
	var initPtr *C.wasm_func_t
	if init != nil {
		initPtr = init.ptr()
	}
	var ptr *C.wasm_table_t
	err := C.wasmtime_funcref_table_new(store.ptr(), ty.ptr(), initPtr, &ptr)
	runtime.KeepAlive(store)
	runtime.KeepAlive(ty)
	runtime.KeepAlive(init)
	if err != nil {
		return nil, mkError(err)
	}
	return mkTable(ptr, store.freelist, nil), nil
}

func mkTable(ptr *C.wasm_table_t, freelist *freeList, owner interface{}) *Table {
	f := &Table{_ptr: ptr, _owner: owner, freelist: freelist}
	if owner == nil {
		runtime.SetFinalizer(f, func(f *Table) {
			f.freelist.lock.Lock()
			defer f.freelist.lock.Unlock()
			f.freelist.tables = append(f.freelist.tables, f._ptr)
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

// Size returns the size of this table in units of elements.
func (t *Table) Size() uint32 {
	ret := C.wasm_table_size(t.ptr())
	runtime.KeepAlive(t)
	return uint32(ret)
}

// Grow grows this funcref table by the number of units specified, using the
// specified initializer value for new slots.
//
// Note that `init` is allowed to be `nil`.
//
// Returns an error if the table failed to grow, or the previous size of the
// table if growth was successful.
func (t *Table) Grow(delta uint32, init *Func) (uint32, error) {
	var initPtr *C.wasm_func_t
	if init != nil {
		initPtr = init.ptr()
	}
	var prev C.uint32_t
	err := C.wasmtime_funcref_table_grow(t.ptr(), C.uint32_t(delta), initPtr, &prev)
	runtime.KeepAlive(t)
	runtime.KeepAlive(init)
	if err == nil {
		return uint32(prev), nil
	}

	return 0, mkError(err)
}

// Get gets an item from this table from the specified index.
//
// Returns an error if the index is out of bounds, or returns a function (which
// may be `nil`) if the index is in bounds corresponding to the entry at the
// specified index.
func (t *Table) Get(idx uint32) (*Func, error) {
	var funcPtr *C.wasm_func_t
	ok := C.wasmtime_funcref_table_get(t.ptr(), C.uint32_t(idx), &funcPtr)
	runtime.KeepAlive(t)
	if ok {
		if funcPtr == nil {
			return nil, nil
		}
		return mkFunc(funcPtr, t.freelist, nil), nil
	}

	return nil, errors.New("index out of bounds")
}

// Set sets an item in this table at the specified index.
//
// Returns an error if the index is out of bounds.
func (t *Table) Set(idx uint32, val *Func) error {
	var funcPtr *C.wasm_func_t
	if val != nil {
		funcPtr = val.ptr()
	}
	err := C.wasmtime_funcref_table_set(t.ptr(), C.uint32_t(idx), funcPtr)
	runtime.KeepAlive(t)
	runtime.KeepAlive(val)
	if err == nil {
		return nil
	}

	return mkError(err)
}

// Type returns the underlying type of this table
func (t *Table) Type() *TableType {
	ptr := C.wasm_table_type(t.ptr())
	runtime.KeepAlive(t)
	return mkTableType(ptr, nil)
}

func (t *Table) AsExtern() *Extern {
	ptr := C.wasm_table_as_extern(t.ptr())
	return mkExtern(ptr, t.freelist, t.owner())
}
