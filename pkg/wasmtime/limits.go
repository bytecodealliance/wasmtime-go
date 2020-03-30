package wasmtime

// #include <wasm.h>
import "C"
import "runtime"

const LIMITS_MAX_NONE = 0xffffffff

type Limits struct {
	Min uint32
	Max uint32
}

func (limits Limits) ffi() C.wasm_limits_t {
	return C.wasm_limits_t{
		min: C.uint32_t(limits.Min),
		max: C.uint32_t(limits.Max),
	}
}

func mkLimits(ptr *C.wasm_limits_t, owner interface{}) Limits {
	ret := Limits{
		Min: uint32(ptr.min),
		Max: uint32(ptr.max),
	}
	runtime.KeepAlive(owner)
	return ret
}
