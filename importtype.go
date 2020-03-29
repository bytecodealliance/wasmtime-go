package wasmtime

// #include <wasm.h>
import "C"
import "runtime"
import "unsafe"

type ImportType struct {
	_ptr   *C.wasm_importtype_t
	_owner interface{}
}

type AsExtern interface {
	AsExtern() *ExternType
}

// Creates a new `ImportType` with the given `module` and `name` and the type
// provided.
func NewImportType(module, name string, ty AsExtern) *ImportType {
        // Convert `module` into an owned `wasm_byte_vec_t`...
	module_vec := C.wasm_byte_vec_t{}
	C.wasm_byte_vec_new_uninitialized(&module_vec, C.size_t(len(module)))
	C.memcpy(unsafe.Pointer(module_vec.data), unsafe.Pointer(C._GoStringPtr(module)), module_vec.size)
	runtime.KeepAlive(module)

        // ... and do the same for `name`
	name_vec := C.wasm_byte_vec_t{}
	C.wasm_byte_vec_new_uninitialized(&name_vec, C.size_t(len(name)))
	C.memcpy(unsafe.Pointer(name_vec.data), unsafe.Pointer(C._GoStringPtr(name)), name_vec.size)
	runtime.KeepAlive(name)

	// Creating an import type requires taking ownership, so create a copy
	// so we don't have to invalidate pointers here. Shouldn't be too
	// costly in theory anyway.
	extern := ty.AsExtern()
	ptr := C.wasm_externtype_copy(extern.ptr())
	runtime.KeepAlive(extern)

	// And once we've got all that create the import type!
	import_ptr := C.wasm_importtype_new(&module_vec, &name_vec, ptr)

	return mkImportType(import_ptr, nil)
}

func mkImportType(ptr *C.wasm_importtype_t, owner interface{}) *ImportType {
	importtype := &ImportType{_ptr: ptr, _owner: owner}
	if owner == nil {
		runtime.SetFinalizer(importtype, func(importtype *ImportType) {
			C.wasm_importtype_delete(importtype._ptr)
		})
	}
	return importtype
}

func (ty *ImportType) ptr() *C.wasm_importtype_t {
	ret := ty._ptr
	maybeGC()
	return ret
}

func (ty *ImportType) owner() interface{} {
	if ty._owner != nil {
		return ty._owner
	}
	return ty
}

// Returns the name in the module this import type is importing
func (ty *ImportType) Module() string {
	ptr := C.wasm_importtype_module(ty.ptr())
	ret := C.GoStringN(ptr.data, C.int(ptr.size))
	runtime.KeepAlive(ty)
	return ret
}

// Returns the name in the module this import type is importing
func (ty *ImportType) Name() string {
	ptr := C.wasm_importtype_name(ty.ptr())
	ret := C.GoStringN(ptr.data, C.int(ptr.size))
	runtime.KeepAlive(ty)
	return ret
}

// Returns the type of item this import type expects
func (ty *ImportType) Type() *ExternType {
	ptr := C.wasm_importtype_type(ty.ptr())
	return mkExternType(ptr, ty.owner())
}
