package wasmtime

// #include <wasm.h>
import "C"
import "runtime"

// Value for the Max field in Limits
const LIMITS_MAX_NONE = 0xffffffff

// Resource limits specified for a TableType and MemoryType
type Limits struct {
	// The minimum size of this resource, in units specified by the resource
	// itself.
	Min uint32
	// The maximum size of this resource, in units specified by the resource
	// itself.
	//
	// A value of LIMITS_MAX_NONE will mean that there is no maximum.
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
