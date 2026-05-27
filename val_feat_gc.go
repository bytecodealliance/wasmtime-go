package wasmtime

// #include <wasm.h>
// #include <wasmtime.h>
//
// extern void goFinalizeExternref(void *env);
//
// static bool go_externref_new(wasmtime_context_t *cx, size_t env, wasmtime_externref_t *ref) {
//   return wasmtime_externref_new(cx, (void*) env, goFinalizeExternref, ref);
// }
//
// static wasmtime_externref_t go_wasmtime_val_externref_get(const wasmtime_val_t *val) {
//   return val->of.externref;
// }
//
// static void go_wasmtime_val_externref_set(wasmtime_val_t *val, wasmtime_externref_t ref) {
//   val->of.externref = ref;
// }
import "C"

import (
	"runtime"
	"sync"
	"unsafe"
)

var gExternrefLock sync.Mutex
var gExternrefMap = make(map[int]interface{})
var gExternrefSlab slab

func init() {
	mkValGC = mkValGCImpl
	valKindGC = valKindGCImpl
	valInitializeGC = valInitializeGCImpl
	initializeExternrefArg = initializeExternrefArgImpl
}

func initializeExternrefArgImpl(store Storelike, val interface{}, dst *C.wasmtime_val_t) Val {
	externref := ValExternref(val)
	externref.initialize(store, dst)
	return externref
}

// ValExternref converts a go value to a externref Val
//
// Using `externref` is a way to pass arbitrary Go data into a WebAssembly
// module for it to store. Later, when you get a `Val`, you can extract the type
// with the `Externref()` method.
func ValExternref(val interface{}) Val {
	return Val{kind: C.WASMTIME_EXTERNREF, val: val}
}

// Externref returns the underlying value if this is an `externref`, or panics.
//
// Note that a null `externref` is returned as `nil`.
func (v Val) Externref() interface{} {
	if v.Kind() != KindExternref {
		panic("not an externref")
	}
	return v.val
}

//export goFinalizeExternref
func goFinalizeExternref(env unsafe.Pointer) {
	idx := int(uintptr(env)) - 1
	gExternrefLock.Lock()
	defer gExternrefLock.Unlock()
	delete(gExternrefMap, idx)
	gExternrefSlab.deallocate(idx)
}

func mkValGCImpl(store Storelike, src *C.wasmtime_val_t) Val {
	if src.kind == C.WASMTIME_EXTERNREF {
		val := C.go_wasmtime_val_externref_get(src)
		if val.store_id == 0 {
			return ValExternref(nil)
		}
		data := C.wasmtime_externref_data(store.Context(), &val)
		runtime.KeepAlive(store)

		gExternrefLock.Lock()
		defer gExternrefLock.Unlock()
		return ValExternref(gExternrefMap[int(uintptr(data))-1])
	}
	panic("failed to get kind of `Val`")
}

func valKindGCImpl(kind C.wasmtime_valkind_t) ValKind {
	if kind == C.WASMTIME_EXTERNREF {
		return KindExternref
	}
	panic("failed to get kind of `Val`")
}

func valInitializeGCImpl(v Val, store Storelike, ptr *C.wasmtime_val_t) {
	if v.kind != C.WASMTIME_EXTERNREF {
		panic("failed to get kind of `Val`")
	}
	// If we have a non-nil value then store it in our global map of all
	// externref values. Otherwise there's nothing for us to do since the
	// `ref` field will already be a nil pointer.
	//
	// Note that we add 1 so all non-null externref values are created with
	// non-null pointers.
	if v.val == nil {
		C.go_wasmtime_val_externref_set(ptr, C.wasmtime_externref_t{})
		return
	}
	gExternrefLock.Lock()
	defer gExternrefLock.Unlock()
	index := gExternrefSlab.allocate()
	gExternrefMap[index] = v.val
	var ref C.wasmtime_externref_t
	ok := C.go_externref_new(store.Context(), C.size_t(index+1), &ref)
	runtime.KeepAlive(store)
	if !ok {
		panic("failed to create an externref")
	}
	C.go_wasmtime_val_externref_set(ptr, ref)
}
