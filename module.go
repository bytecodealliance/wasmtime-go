package wasmtime

// #include <wasmtime.h>
//
// wasmtime_error_t *go_module_new(wasm_store_t *store, uint8_t *bytes, size_t len, wasm_module_t **ret) {
//    wasm_byte_vec_t vec;
//    vec.data = (wasm_byte_t*) bytes;
//    vec.size = len;
//    return wasmtime_module_new(store, &vec, ret);
// }
//
// wasmtime_error_t *go_module_validate(wasm_store_t *store, uint8_t *bytes, size_t len) {
//    wasm_byte_vec_t vec;
//    vec.data = (wasm_byte_t*) bytes;
//    vec.size = len;
//    return wasmtime_module_validate(store, &vec);
// }
import "C"
import (
	"io/ioutil"
	"runtime"
	"unsafe"
)

type Module struct {
	_ptr  *C.wasm_module_t
	Store *Store
}

// Compiles a new `Module` from the `wasm` provided with the given configuration
// in `store`.
func NewModule(store *Store, wasm []byte) (*Module, error) {
	// We can't create the `wasm_byte_vec_t` here and pass it in because
	// that runs into the error of "passed a pointer to a pointer" because
	// the vec itself is passed by pointer and it contains a pointer to
	// `wasm`. To work around this we insert some C shims above and call
	// them.
	var wasm_ptr *C.uint8_t
	if len(wasm) > 0 {
		wasm_ptr = (*C.uint8_t)(unsafe.Pointer(&wasm[0]))
	}
	var ptr *C.wasm_module_t
	err := C.go_module_new(store.ptr(), wasm_ptr, C.size_t(len(wasm)), &ptr)
	runtime.KeepAlive(store)
	runtime.KeepAlive(wasm)

	if err != nil {
		return nil, mkError(err)
	} else {
		return mkModule(ptr, store), nil
	}
}

// Reads the contents of the `file` provided and interprets them as either the
// text format or the binary format for WebAssembly.
//
// Afterwards delegates to the `NewModule` constructor with the contents read.
func NewModuleFromFile(store *Store, file string) (*Module, error) {
	wasm, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	// If this wasm isn't actually wasm, treat it as the text format and
	// parse it as such.
	if len(wasm) > 0 && wasm[0] != 0 {
		wasm, err = Wat2Wasm(string(wasm))
		if err != nil {
			return nil, err
		}
	}
	return NewModule(store, wasm)

}

// Validates whether `wasm` would be a valid wasm module according to the
// configuration in `store`
func ModuleValidate(store *Store, wasm []byte) error {
	var wasm_ptr *C.uint8_t
	if len(wasm) > 0 {
		wasm_ptr = (*C.uint8_t)(unsafe.Pointer(&wasm[0]))
	}
	err := C.go_module_validate(store.ptr(), wasm_ptr, C.size_t(len(wasm)))
	runtime.KeepAlive(store)
	runtime.KeepAlive(wasm)
	if err == nil {
		return nil
	} else {
		return mkError(err)
	}
}

func mkModule(ptr *C.wasm_module_t, store *Store) *Module {
	module := &Module{_ptr: ptr, Store: store}
	runtime.SetFinalizer(module, func(module *Module) {
		freelist := module.Store.freelist
		freelist.lock.Lock()
		defer freelist.lock.Unlock()
		freelist.modules = append(freelist.modules, module._ptr)
	})
	return module
}

func (m *Module) ptr() *C.wasm_module_t {
	ret := m._ptr
	maybeGC()
	return ret
}

type importTypeList struct {
	vec C.wasm_importtype_vec_t
}

// Returns a list of `ImportType` items which are the items imported by this
// module and are required for instantiation.
func (m *Module) Imports() []*ImportType {
	imports := &importTypeList{}
	C.wasm_module_imports(m.ptr(), &imports.vec)
	runtime.KeepAlive(m)
	runtime.SetFinalizer(imports, func(imports *importTypeList) {
		C.wasm_importtype_vec_delete(&imports.vec)
	})

	ret := make([]*ImportType, int(imports.vec.size))
	base := unsafe.Pointer(imports.vec.data)
	var ptr *C.wasm_importtype_t
	for i := 0; i < int(imports.vec.size); i++ {
		ptr := *(**C.wasm_importtype_t)(unsafe.Pointer(uintptr(base) + unsafe.Sizeof(ptr)*uintptr(i)))
		ty := mkImportType(ptr, imports)
		ret[i] = ty
	}
	return ret
}

type exportTypeList struct {
	vec C.wasm_exporttype_vec_t
}

// Returns a list of `ExportType` items which are the items that will
// be exported by this module after instantiation.
func (m *Module) Exports() []*ExportType {
	exports := &exportTypeList{}
	C.wasm_module_exports(m.ptr(), &exports.vec)
	runtime.KeepAlive(m)
	runtime.SetFinalizer(exports, func(exports *exportTypeList) {
		C.wasm_exporttype_vec_delete(&exports.vec)
	})

	ret := make([]*ExportType, int(exports.vec.size))
	base := unsafe.Pointer(exports.vec.data)
	var ptr *C.wasm_exporttype_t
	for i := 0; i < int(exports.vec.size); i++ {
		ptr := *(**C.wasm_exporttype_t)(unsafe.Pointer(uintptr(base) + unsafe.Sizeof(ptr)*uintptr(i)))
		ty := mkExportType(ptr, exports)
		ret[i] = ty
	}
	return ret
}
