//go:build wasmtime_component_model

package wasmtime

/*
#include <wasmtime.h>
#include <stdlib.h>
#include "component_shims_feat_component_model.h"

static inline const wasmtime_component_val_t *go_component_callback_arg_at(const wasmtime_component_val_t *p, size_t i) { return &p[i]; }
static inline wasmtime_component_val_t *go_component_callback_result_at(wasmtime_component_val_t *p, size_t i) { return &p[i]; }
*/
import "C"

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

// ComponentCaller is the store context provided to a host component function.
// It is valid only for the duration of the callback.
type ComponentCaller struct {
	context *C.wasmtime_context_t
}

func (c *ComponentCaller) Context() *C.wasmtime_context_t {
	if c.context == nil {
		panic("component caller is no longer valid")
	}
	return c.context
}

func (c *ComponentCaller) Data() interface{} { return getDataInStore(c).data }

// ComponentHostFunc implements one host-provided component function. Arguments
// are owned clones valid only during the callback. Result ownership transfers
// to the binding and results are closed after they are copied into Wasmtime.
type ComponentHostFunc func(*ComponentCaller, []*ComponentVal) ([]*ComponentVal, error)

var (
	gComponentFuncLock sync.Mutex
	gComponentFuncs    = make(map[int]ComponentHostFunc)
	gComponentFuncSlab slab
)

// ComponentResourceDestructor runs when an owned host resource is dropped.
type ComponentResourceDestructor func(*ComponentCaller, uint32) error

var (
	gComponentResourceLock        sync.Mutex
	gComponentResourceDestructors = make(map[int]ComponentResourceDestructor)
	gComponentResourceSlab        slab
)

func insertComponentResourceDestructor(callback ComponentResourceDestructor) int {
	gComponentResourceLock.Lock()
	defer gComponentResourceLock.Unlock()
	index := gComponentResourceSlab.allocate()
	gComponentResourceDestructors[index] = callback
	return index
}

func removeComponentResourceDestructor(index int) {
	gComponentResourceLock.Lock()
	defer gComponentResourceLock.Unlock()
	delete(gComponentResourceDestructors, index)
	gComponentResourceSlab.deallocate(index)
}

func insertComponentFunc(callback ComponentHostFunc) int {
	gComponentFuncLock.Lock()
	defer gComponentFuncLock.Unlock()
	index := gComponentFuncSlab.allocate()
	gComponentFuncs[index] = callback
	return index
}

func getComponentFunc(index int) ComponentHostFunc {
	gComponentFuncLock.Lock()
	defer gComponentFuncLock.Unlock()
	return gComponentFuncs[index]
}

func removeComponentFunc(index int) {
	gComponentFuncLock.Lock()
	defer gComponentFuncLock.Unlock()
	delete(gComponentFuncs, index)
	gComponentFuncSlab.deallocate(index)
}

//export goFinalizeComponentFunc
func goFinalizeComponentFunc(env unsafe.Pointer) {
	removeComponentFunc(int(uintptr(env)))
}

//export goFinalizeComponentResourceDestructor
func goFinalizeComponentResourceDestructor(env unsafe.Pointer) {
	removeComponentResourceDestructor(int(uintptr(env)))
}

//export goComponentResourceDestructor
func goComponentResourceDestructor(env unsafe.Pointer, context *C.wasmtime_context_t, rep C.uint32_t) (ret *C.wasmtime_error_t) {
	caller := &ComponentCaller{context: context}
	defer func() { caller.context = nil }()
	defer func() {
		if recovered := recover(); recovered != nil {
			getDataInStore(caller).lastPanic = recovered
			ret = componentCallbackError(fmt.Errorf("go component resource destructor panicked"))
		}
	}()
	gComponentResourceLock.Lock()
	callback := gComponentResourceDestructors[int(uintptr(env))]
	gComponentResourceLock.Unlock()
	if callback == nil {
		return componentCallbackError(fmt.Errorf("component resource destructor is unavailable"))
	}
	if err := callback(caller, uint32(rep)); err != nil {
		return componentCallbackError(err)
	}
	return nil
}

func componentCallbackError(err error) *C.wasmtime_error_t {
	message := C.CString(err.Error())
	defer C.free(unsafe.Pointer(message))
	return C.wasmtime_error_new(message)
}

//export goComponentFuncCallback
func goComponentFuncCallback(
	env unsafe.Pointer,
	context *C.wasmtime_context_t,
	_ *C.wasmtime_component_func_type_t,
	args *C.wasmtime_component_val_t,
	argsLen C.size_t,
	results *C.wasmtime_component_val_t,
	resultsLen C.size_t,
) (ret *C.wasmtime_error_t) {
	caller := &ComponentCaller{context: context}
	defer func() { caller.context = nil }()
	defer func() {
		if recovered := recover(); recovered != nil {
			getDataInStore(caller).lastPanic = recovered
			ret = componentCallbackError(fmt.Errorf("go component host function panicked"))
		}
	}()

	callback := getComponentFunc(int(uintptr(env)))
	if callback == nil {
		return componentCallbackError(fmt.Errorf("component host function callback is unavailable"))
	}
	arguments := make([]*ComponentVal, int(argsLen))
	for i := range arguments {
		var cloned C.wasmtime_component_val_t
		C.wasmtime_component_val_clone(C.go_component_callback_arg_at(args, C.size_t(i)), &cloned)
		arguments[i] = ownComponentVal(cloned)
	}
	defer closeComponentValues(arguments)

	returned, err := callback(caller, arguments)
	if err != nil {
		closeComponentValues(returned)
		return componentCallbackError(err)
	}
	defer closeComponentValues(returned)
	if len(returned) != int(resultsLen) {
		return componentCallbackError(fmt.Errorf("component host function returned %d values, want %d", len(returned), int(resultsLen)))
	}
	for i, value := range returned {
		C.wasmtime_component_val_clone(value.ptr(), C.go_component_callback_result_at(results, C.size_t(i)))
	}
	runtime.KeepAlive(returned)
	return nil
}

func closeComponentValues(values []*ComponentVal) {
	for _, value := range values {
		if value != nil {
			value.Close()
		}
	}
}
