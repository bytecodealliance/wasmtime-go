package wasmtime

// #include <wasmtime.h>
// #include <string.h>
//
// static inline void go_component_val_set_bool(wasmtime_component_val_t *v, bool x) { v->of.boolean = x; }
// static inline bool go_component_val_get_bool(const wasmtime_component_val_t *v) { return v->of.boolean; }
//
// static inline void go_component_val_set_s8(wasmtime_component_val_t *v, int8_t x) { v->of.s8 = x; }
// static inline int8_t go_component_val_get_s8(const wasmtime_component_val_t *v) { return v->of.s8; }
//
// static inline void go_component_val_set_u8(wasmtime_component_val_t *v, uint8_t x) { v->of.u8 = x; }
// static inline uint8_t go_component_val_get_u8(const wasmtime_component_val_t *v) { return v->of.u8; }
//
// static inline void go_component_val_set_s16(wasmtime_component_val_t *v, int16_t x) { v->of.s16 = x; }
// static inline int16_t go_component_val_get_s16(const wasmtime_component_val_t *v) { return v->of.s16; }
//
// static inline void go_component_val_set_u16(wasmtime_component_val_t *v, uint16_t x) { v->of.u16 = x; }
// static inline uint16_t go_component_val_get_u16(const wasmtime_component_val_t *v) { return v->of.u16; }
//
// static inline void go_component_val_set_s32(wasmtime_component_val_t *v, int32_t x) { v->of.s32 = x; }
// static inline int32_t go_component_val_get_s32(const wasmtime_component_val_t *v) { return v->of.s32; }
//
// static inline void go_component_val_set_u32(wasmtime_component_val_t *v, uint32_t x) { v->of.u32 = x; }
// static inline uint32_t go_component_val_get_u32(const wasmtime_component_val_t *v) { return v->of.u32; }
//
// static inline void go_component_val_set_s64(wasmtime_component_val_t *v, int64_t x) { v->of.s64 = x; }
// static inline int64_t go_component_val_get_s64(const wasmtime_component_val_t *v) { return v->of.s64; }
//
// static inline void go_component_val_set_u64(wasmtime_component_val_t *v, uint64_t x) { v->of.u64 = x; }
// static inline uint64_t go_component_val_get_u64(const wasmtime_component_val_t *v) { return v->of.u64; }
//
// static inline void go_component_val_set_f32(wasmtime_component_val_t *v, float x) { v->of.f32 = x; }
// static inline float go_component_val_get_f32(const wasmtime_component_val_t *v) { return v->of.f32; }
//
// static inline void go_component_val_set_f64(wasmtime_component_val_t *v, double x) { v->of.f64 = x; }
// static inline double go_component_val_get_f64(const wasmtime_component_val_t *v) { return v->of.f64; }
//
// static inline void go_component_val_set_char(wasmtime_component_val_t *v, uint32_t x) { v->of.character = x; }
// static inline uint32_t go_component_val_get_char(const wasmtime_component_val_t *v) { return v->of.character; }
//
// static inline void go_component_val_set_string(wasmtime_component_val_t *v, const char *data, size_t len) {
//   wasm_byte_vec_new_uninitialized(&v->of.string, len);
//   if (len > 0) memcpy(v->of.string.data, data, len);
// }
// static inline const char *go_component_val_string_data(const wasmtime_component_val_t *v) { return v->of.string.data; }
// static inline size_t go_component_val_string_size(const wasmtime_component_val_t *v) { return v->of.string.size; }
import "C"

import (
	"fmt"
	"runtime"
)

// componentMarshalArg writes the Go value `arg` into the C-side `out` slot,
// matching the WIT type discriminated by `kind` (a valtype kind, not a val
// kind). Only primitive WIT types are supported in this version.
func componentMarshalArg(arg interface{}, kind C.wasmtime_component_valtype_kind_t, out *C.wasmtime_component_val_t) error {
	switch kind {
	case C.WASMTIME_COMPONENT_VALTYPE_BOOL:
		v, ok := arg.(bool)
		if !ok {
			return componentArgMismatch("bool", arg)
		}
		out.kind = C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_BOOL)
		C.go_component_val_set_bool(out, C.bool(v))
	case C.WASMTIME_COMPONENT_VALTYPE_S8:
		v, ok := arg.(int8)
		if !ok {
			return componentArgMismatch("int8", arg)
		}
		out.kind = C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_S8)
		C.go_component_val_set_s8(out, C.int8_t(v))
	case C.WASMTIME_COMPONENT_VALTYPE_U8:
		v, ok := arg.(uint8)
		if !ok {
			return componentArgMismatch("uint8", arg)
		}
		out.kind = C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_U8)
		C.go_component_val_set_u8(out, C.uint8_t(v))
	case C.WASMTIME_COMPONENT_VALTYPE_S16:
		v, ok := arg.(int16)
		if !ok {
			return componentArgMismatch("int16", arg)
		}
		out.kind = C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_S16)
		C.go_component_val_set_s16(out, C.int16_t(v))
	case C.WASMTIME_COMPONENT_VALTYPE_U16:
		v, ok := arg.(uint16)
		if !ok {
			return componentArgMismatch("uint16", arg)
		}
		out.kind = C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_U16)
		C.go_component_val_set_u16(out, C.uint16_t(v))
	case C.WASMTIME_COMPONENT_VALTYPE_S32:
		v, ok := arg.(int32)
		if !ok {
			return componentArgMismatch("int32", arg)
		}
		out.kind = C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_S32)
		C.go_component_val_set_s32(out, C.int32_t(v))
	case C.WASMTIME_COMPONENT_VALTYPE_U32:
		v, ok := arg.(uint32)
		if !ok {
			return componentArgMismatch("uint32", arg)
		}
		out.kind = C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_U32)
		C.go_component_val_set_u32(out, C.uint32_t(v))
	case C.WASMTIME_COMPONENT_VALTYPE_S64:
		v, ok := arg.(int64)
		if !ok {
			return componentArgMismatch("int64", arg)
		}
		out.kind = C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_S64)
		C.go_component_val_set_s64(out, C.int64_t(v))
	case C.WASMTIME_COMPONENT_VALTYPE_U64:
		v, ok := arg.(uint64)
		if !ok {
			return componentArgMismatch("uint64", arg)
		}
		out.kind = C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_U64)
		C.go_component_val_set_u64(out, C.uint64_t(v))
	case C.WASMTIME_COMPONENT_VALTYPE_F32:
		v, ok := arg.(float32)
		if !ok {
			return componentArgMismatch("float32", arg)
		}
		out.kind = C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_F32)
		C.go_component_val_set_f32(out, C.float(v))
	case C.WASMTIME_COMPONENT_VALTYPE_F64:
		v, ok := arg.(float64)
		if !ok {
			return componentArgMismatch("float64", arg)
		}
		out.kind = C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_F64)
		C.go_component_val_set_f64(out, C.double(v))
	case C.WASMTIME_COMPONENT_VALTYPE_CHAR:
		// `rune` is an alias for `int32`, so users may pass either.
		v, ok := arg.(int32)
		if !ok {
			return componentArgMismatch("rune (int32)", arg)
		}
		out.kind = C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_CHAR)
		C.go_component_val_set_char(out, C.uint32_t(v))
	case C.WASMTIME_COMPONENT_VALTYPE_STRING:
		v, ok := arg.(string)
		if !ok {
			return componentArgMismatch("string", arg)
		}
		out.kind = C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_STRING)
		C.go_component_val_set_string(out, C._GoStringPtr(v), C._GoStringLen(v))
		runtime.KeepAlive(v)
	default:
		return fmt.Errorf("unsupported component type kind: %d (only primitive WIT types are supported in this version)", kind)
	}
	return nil
}

// componentUnmarshalVal converts a C-side `wasmtime_component_val_t` into a
// Go value. Only primitive WIT types are supported in this version.
func componentUnmarshalVal(v *C.wasmtime_component_val_t) (interface{}, error) {
	switch v.kind {
	case C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_BOOL):
		return bool(C.go_component_val_get_bool(v)), nil
	case C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_S8):
		return int8(C.go_component_val_get_s8(v)), nil
	case C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_U8):
		return uint8(C.go_component_val_get_u8(v)), nil
	case C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_S16):
		return int16(C.go_component_val_get_s16(v)), nil
	case C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_U16):
		return uint16(C.go_component_val_get_u16(v)), nil
	case C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_S32):
		return int32(C.go_component_val_get_s32(v)), nil
	case C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_U32):
		return uint32(C.go_component_val_get_u32(v)), nil
	case C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_S64):
		return int64(C.go_component_val_get_s64(v)), nil
	case C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_U64):
		return uint64(C.go_component_val_get_u64(v)), nil
	case C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_F32):
		return float32(C.go_component_val_get_f32(v)), nil
	case C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_F64):
		return float64(C.go_component_val_get_f64(v)), nil
	case C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_CHAR):
		return rune(C.go_component_val_get_char(v)), nil
	case C.wasmtime_component_valkind_t(C.WASMTIME_COMPONENT_STRING):
		data := C.go_component_val_string_data(v)
		size := C.go_component_val_string_size(v)
		return C.GoStringN(data, C.int(size)), nil
	default:
		return nil, fmt.Errorf("unsupported component value kind: %d (only primitive WIT types are supported in this version)", v.kind)
	}
}

func componentArgMismatch(expected string, got interface{}) error {
	return fmt.Errorf("type mismatch: expected %s, got %T", expected, got)
}
