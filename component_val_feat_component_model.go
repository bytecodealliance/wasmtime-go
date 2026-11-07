//go:build wasmtime_component_model

package wasmtime

/*
#include <wasmtime.h>
#include <stdlib.h>

static inline uint8_t go_component_val_kind(const wasmtime_component_val_t *v) { return v->kind; }
static inline void go_component_val_bool(wasmtime_component_val_t *v, bool x) { v->kind = WASMTIME_COMPONENT_BOOL; v->of.boolean = x; }
static inline void go_component_val_s8(wasmtime_component_val_t *v, int8_t x) { v->kind = WASMTIME_COMPONENT_S8; v->of.s8 = x; }
static inline void go_component_val_u8(wasmtime_component_val_t *v, uint8_t x) { v->kind = WASMTIME_COMPONENT_U8; v->of.u8 = x; }
static inline void go_component_val_s16(wasmtime_component_val_t *v, int16_t x) { v->kind = WASMTIME_COMPONENT_S16; v->of.s16 = x; }
static inline void go_component_val_u16(wasmtime_component_val_t *v, uint16_t x) { v->kind = WASMTIME_COMPONENT_U16; v->of.u16 = x; }
static inline void go_component_val_s32(wasmtime_component_val_t *v, int32_t x) { v->kind = WASMTIME_COMPONENT_S32; v->of.s32 = x; }
static inline void go_component_val_u32(wasmtime_component_val_t *v, uint32_t x) { v->kind = WASMTIME_COMPONENT_U32; v->of.u32 = x; }
static inline void go_component_val_s64(wasmtime_component_val_t *v, int64_t x) { v->kind = WASMTIME_COMPONENT_S64; v->of.s64 = x; }
static inline void go_component_val_u64(wasmtime_component_val_t *v, uint64_t x) { v->kind = WASMTIME_COMPONENT_U64; v->of.u64 = x; }
static inline void go_component_val_f32(wasmtime_component_val_t *v, float x) { v->kind = WASMTIME_COMPONENT_F32; v->of.f32 = x; }
static inline void go_component_val_f64(wasmtime_component_val_t *v, double x) { v->kind = WASMTIME_COMPONENT_F64; v->of.f64 = x; }
static inline void go_component_val_char(wasmtime_component_val_t *v, uint32_t x) { v->kind = WASMTIME_COMPONENT_CHAR; v->of.character = x; }
static inline void go_component_val_string(wasmtime_component_val_t *v, const char *p, size_t n) { v->kind = WASMTIME_COMPONENT_STRING; wasm_name_new(&v->of.string, n, p); }
static inline void go_component_val_enum(wasmtime_component_val_t *v, const char *p, size_t n) { v->kind = WASMTIME_COMPONENT_ENUM; wasm_name_new(&v->of.enumeration, n, p); }

static inline bool go_component_val_get_bool(const wasmtime_component_val_t *v) { return v->of.boolean; }
static inline int8_t go_component_val_get_s8(const wasmtime_component_val_t *v) { return v->of.s8; }
static inline uint8_t go_component_val_get_u8(const wasmtime_component_val_t *v) { return v->of.u8; }
static inline int16_t go_component_val_get_s16(const wasmtime_component_val_t *v) { return v->of.s16; }
static inline uint16_t go_component_val_get_u16(const wasmtime_component_val_t *v) { return v->of.u16; }
static inline int32_t go_component_val_get_s32(const wasmtime_component_val_t *v) { return v->of.s32; }
static inline uint32_t go_component_val_get_u32(const wasmtime_component_val_t *v) { return v->of.u32; }
static inline int64_t go_component_val_get_s64(const wasmtime_component_val_t *v) { return v->of.s64; }
static inline uint64_t go_component_val_get_u64(const wasmtime_component_val_t *v) { return v->of.u64; }
static inline float go_component_val_get_f32(const wasmtime_component_val_t *v) { return v->of.f32; }
static inline double go_component_val_get_f64(const wasmtime_component_val_t *v) { return v->of.f64; }
static inline uint32_t go_component_val_get_char(const wasmtime_component_val_t *v) { return v->of.character; }
static inline const char *go_component_val_get_string(const wasmtime_component_val_t *v) { return v->of.string.data; }
static inline size_t go_component_val_get_string_len(const wasmtime_component_val_t *v) { return v->of.string.size; }
static inline const char *go_component_val_get_enum(const wasmtime_component_val_t *v) { return v->of.enumeration.data; }
static inline size_t go_component_val_get_enum_len(const wasmtime_component_val_t *v) { return v->of.enumeration.size; }

static inline void go_component_val_list_init(wasmtime_component_val_t *v, size_t n) { v->kind = WASMTIME_COMPONENT_LIST; wasmtime_component_vallist_new_uninit(&v->of.list, n); }
static inline void go_component_val_list_set(wasmtime_component_val_t *v, size_t i, const wasmtime_component_val_t *x) { wasmtime_component_val_clone(x, &v->of.list.data[i]); }
static inline size_t go_component_val_list_len(const wasmtime_component_val_t *v) { return v->of.list.size; }
static inline void go_component_val_list_get(const wasmtime_component_val_t *v, size_t i, wasmtime_component_val_t *out) { wasmtime_component_val_clone(&v->of.list.data[i], out); }

static inline void go_component_val_tuple_init(wasmtime_component_val_t *v, size_t n) { v->kind = WASMTIME_COMPONENT_TUPLE; wasmtime_component_valtuple_new_uninit(&v->of.tuple, n); }
static inline void go_component_val_tuple_set(wasmtime_component_val_t *v, size_t i, const wasmtime_component_val_t *x) { wasmtime_component_val_clone(x, &v->of.tuple.data[i]); }
static inline size_t go_component_val_tuple_len(const wasmtime_component_val_t *v) { return v->of.tuple.size; }
static inline void go_component_val_tuple_get(const wasmtime_component_val_t *v, size_t i, wasmtime_component_val_t *out) { wasmtime_component_val_clone(&v->of.tuple.data[i], out); }

static inline void go_component_val_record_init(wasmtime_component_val_t *v, size_t n) { v->kind = WASMTIME_COMPONENT_RECORD; wasmtime_component_valrecord_new_uninit(&v->of.record, n); }
static inline void go_component_val_record_set(wasmtime_component_val_t *v, size_t i, const char *p, size_t n, const wasmtime_component_val_t *x) { wasm_name_new(&v->of.record.data[i].name, n, p); wasmtime_component_val_clone(x, &v->of.record.data[i].val); }
static inline size_t go_component_val_record_len(const wasmtime_component_val_t *v) { return v->of.record.size; }
static inline const char *go_component_val_record_name(const wasmtime_component_val_t *v, size_t i) { return v->of.record.data[i].name.data; }
static inline size_t go_component_val_record_name_len(const wasmtime_component_val_t *v, size_t i) { return v->of.record.data[i].name.size; }
static inline void go_component_val_record_get(const wasmtime_component_val_t *v, size_t i, wasmtime_component_val_t *out) { wasmtime_component_val_clone(&v->of.record.data[i].val, out); }

static inline void go_component_val_variant(wasmtime_component_val_t *v, const char *p, size_t n, const wasmtime_component_val_t *x) { v->kind = WASMTIME_COMPONENT_VARIANT; wasm_name_new(&v->of.variant.discriminant, n, p); if (x == NULL) { v->of.variant.val = NULL; } else { wasmtime_component_val_t tmp; wasmtime_component_val_clone(x, &tmp); v->of.variant.val = wasmtime_component_val_new(&tmp); } }
static inline const char *go_component_val_variant_name(const wasmtime_component_val_t *v) { return v->of.variant.discriminant.data; }
static inline size_t go_component_val_variant_name_len(const wasmtime_component_val_t *v) { return v->of.variant.discriminant.size; }
static inline bool go_component_val_variant_get(const wasmtime_component_val_t *v, wasmtime_component_val_t *out) { if (v->of.variant.val == NULL) return false; wasmtime_component_val_clone(v->of.variant.val, out); return true; }

static inline void go_component_val_option(wasmtime_component_val_t *v, const wasmtime_component_val_t *x) { v->kind = WASMTIME_COMPONENT_OPTION; if (x == NULL) { v->of.option = NULL; } else { wasmtime_component_val_t tmp; wasmtime_component_val_clone(x, &tmp); v->of.option = wasmtime_component_val_new(&tmp); } }
static inline bool go_component_val_option_get(const wasmtime_component_val_t *v, wasmtime_component_val_t *out) { if (v->of.option == NULL) return false; wasmtime_component_val_clone(v->of.option, out); return true; }

static inline void go_component_val_result(wasmtime_component_val_t *v, bool ok, const wasmtime_component_val_t *x) { v->kind = WASMTIME_COMPONENT_RESULT; v->of.result.is_ok = ok; if (x == NULL) { v->of.result.val = NULL; } else { wasmtime_component_val_t tmp; wasmtime_component_val_clone(x, &tmp); v->of.result.val = wasmtime_component_val_new(&tmp); } }
static inline bool go_component_val_result_ok(const wasmtime_component_val_t *v) { return v->of.result.is_ok; }
static inline bool go_component_val_result_get(const wasmtime_component_val_t *v, wasmtime_component_val_t *out) { if (v->of.result.val == NULL) return false; wasmtime_component_val_clone(v->of.result.val, out); return true; }

static inline void go_component_val_flags_init(wasmtime_component_val_t *v, size_t n) { v->kind = WASMTIME_COMPONENT_FLAGS; wasmtime_component_valflags_new_uninit(&v->of.flags, n); }
static inline void go_component_val_flags_set(wasmtime_component_val_t *v, size_t i, const char *p, size_t n) { wasm_name_new(&v->of.flags.data[i], n, p); }
static inline size_t go_component_val_flags_len(const wasmtime_component_val_t *v) { return v->of.flags.size; }
static inline const char *go_component_val_flags_get(const wasmtime_component_val_t *v, size_t i) { return v->of.flags.data[i].data; }
static inline size_t go_component_val_flags_get_len(const wasmtime_component_val_t *v, size_t i) { return v->of.flags.data[i].size; }
*/
import "C"

import (
	"runtime"
	"unsafe"
)

// ComponentValKind identifies the WIT runtime value stored in a ComponentVal.
type ComponentValKind uint8

const (
	ComponentValKindBool    ComponentValKind = C.WASMTIME_COMPONENT_BOOL
	ComponentValKindS8      ComponentValKind = C.WASMTIME_COMPONENT_S8
	ComponentValKindU8      ComponentValKind = C.WASMTIME_COMPONENT_U8
	ComponentValKindS16     ComponentValKind = C.WASMTIME_COMPONENT_S16
	ComponentValKindU16     ComponentValKind = C.WASMTIME_COMPONENT_U16
	ComponentValKindS32     ComponentValKind = C.WASMTIME_COMPONENT_S32
	ComponentValKindU32     ComponentValKind = C.WASMTIME_COMPONENT_U32
	ComponentValKindS64     ComponentValKind = C.WASMTIME_COMPONENT_S64
	ComponentValKindU64     ComponentValKind = C.WASMTIME_COMPONENT_U64
	ComponentValKindF32     ComponentValKind = C.WASMTIME_COMPONENT_F32
	ComponentValKindF64     ComponentValKind = C.WASMTIME_COMPONENT_F64
	ComponentValKindChar    ComponentValKind = C.WASMTIME_COMPONENT_CHAR
	ComponentValKindString  ComponentValKind = C.WASMTIME_COMPONENT_STRING
	ComponentValKindList    ComponentValKind = C.WASMTIME_COMPONENT_LIST
	ComponentValKindRecord  ComponentValKind = C.WASMTIME_COMPONENT_RECORD
	ComponentValKindTuple   ComponentValKind = C.WASMTIME_COMPONENT_TUPLE
	ComponentValKindVariant ComponentValKind = C.WASMTIME_COMPONENT_VARIANT
	ComponentValKindEnum    ComponentValKind = C.WASMTIME_COMPONENT_ENUM
	ComponentValKindOption  ComponentValKind = C.WASMTIME_COMPONENT_OPTION
	ComponentValKindResult  ComponentValKind = C.WASMTIME_COMPONENT_RESULT
	ComponentValKindFlags   ComponentValKind = C.WASMTIME_COMPONENT_FLAGS
)

// ComponentVal is an owned dynamically typed component-model value.
// Close it explicitly when it is no longer needed.
type ComponentVal struct {
	val    C.wasmtime_component_val_t
	closed bool
}

// ComponentRecordField is one named field of a record value.
type ComponentRecordField struct {
	Name  string
	Value *ComponentVal
}

// ComponentVariantValue is a variant discriminant and optional payload.
type ComponentVariantValue struct {
	Discriminant string
	Value        *ComponentVal
}

// ComponentResultValue is an ok/error result and optional payload.
type ComponentResultValue struct {
	OK    bool
	Value *ComponentVal
}

func ownComponentVal(val C.wasmtime_component_val_t) *ComponentVal {
	v := &ComponentVal{val: val}
	runtime.SetFinalizer(v, func(v *ComponentVal) { v.Close() })
	return v
}

func newComponentVal(set func(*C.wasmtime_component_val_t)) *ComponentVal {
	v := ownComponentVal(C.wasmtime_component_val_t{})
	set(&v.val)
	return v
}

func NewComponentBool(x bool) *ComponentVal {
	return newComponentVal(func(v *C.wasmtime_component_val_t) { C.go_component_val_bool(v, C.bool(x)) })
}

func NewComponentS8(x int8) *ComponentVal {
	return newComponentVal(func(v *C.wasmtime_component_val_t) { C.go_component_val_s8(v, C.int8_t(x)) })
}

func NewComponentU8(x uint8) *ComponentVal {
	return newComponentVal(func(v *C.wasmtime_component_val_t) { C.go_component_val_u8(v, C.uint8_t(x)) })
}

func NewComponentS16(x int16) *ComponentVal {
	return newComponentVal(func(v *C.wasmtime_component_val_t) { C.go_component_val_s16(v, C.int16_t(x)) })
}

func NewComponentU16(x uint16) *ComponentVal {
	return newComponentVal(func(v *C.wasmtime_component_val_t) { C.go_component_val_u16(v, C.uint16_t(x)) })
}

func NewComponentS32(x int32) *ComponentVal {
	return newComponentVal(func(v *C.wasmtime_component_val_t) { C.go_component_val_s32(v, C.int32_t(x)) })
}

func NewComponentU32(x uint32) *ComponentVal {
	return newComponentVal(func(v *C.wasmtime_component_val_t) { C.go_component_val_u32(v, C.uint32_t(x)) })
}

func NewComponentS64(x int64) *ComponentVal {
	return newComponentVal(func(v *C.wasmtime_component_val_t) { C.go_component_val_s64(v, C.int64_t(x)) })
}

func NewComponentU64(x uint64) *ComponentVal {
	return newComponentVal(func(v *C.wasmtime_component_val_t) { C.go_component_val_u64(v, C.uint64_t(x)) })
}

func NewComponentF32(x float32) *ComponentVal {
	return newComponentVal(func(v *C.wasmtime_component_val_t) { C.go_component_val_f32(v, C.float(x)) })
}

func NewComponentF64(x float64) *ComponentVal {
	return newComponentVal(func(v *C.wasmtime_component_val_t) { C.go_component_val_f64(v, C.double(x)) })
}

func NewComponentChar(x rune) *ComponentVal {
	return newComponentVal(func(v *C.wasmtime_component_val_t) { C.go_component_val_char(v, C.uint32_t(x)) })
}

func NewComponentString(x string) *ComponentVal {
	return newComponentVal(func(v *C.wasmtime_component_val_t) {
		C.go_component_val_string(v, C._GoStringPtr(x), C._GoStringLen(x))
	})
}

func NewComponentEnum(x string) *ComponentVal {
	return newComponentVal(func(v *C.wasmtime_component_val_t) { C.go_component_val_enum(v, C._GoStringPtr(x), C._GoStringLen(x)) })
}

func NewComponentList(values []*ComponentVal) *ComponentVal {
	v := newComponentVal(func(v *C.wasmtime_component_val_t) { C.go_component_val_list_init(v, C.size_t(len(values))) })
	for i, value := range values {
		C.go_component_val_list_set(&v.val, C.size_t(i), value.ptr())
	}
	runtime.KeepAlive(values)
	return v
}

func NewComponentTuple(values []*ComponentVal) *ComponentVal {
	v := newComponentVal(func(v *C.wasmtime_component_val_t) { C.go_component_val_tuple_init(v, C.size_t(len(values))) })
	for i, value := range values {
		C.go_component_val_tuple_set(&v.val, C.size_t(i), value.ptr())
	}
	runtime.KeepAlive(values)
	return v
}

func NewComponentRecord(fields []ComponentRecordField) *ComponentVal {
	v := newComponentVal(func(v *C.wasmtime_component_val_t) { C.go_component_val_record_init(v, C.size_t(len(fields))) })
	for i, field := range fields {
		C.go_component_val_record_set(&v.val, C.size_t(i), C._GoStringPtr(field.Name), C._GoStringLen(field.Name), field.Value.ptr())
	}
	runtime.KeepAlive(fields)
	return v
}

func NewComponentVariant(discriminant string, value *ComponentVal) *ComponentVal {
	return newComponentVal(func(v *C.wasmtime_component_val_t) {
		var payload *C.wasmtime_component_val_t
		if value != nil {
			payload = value.ptr()
		}
		C.go_component_val_variant(v, C._GoStringPtr(discriminant), C._GoStringLen(discriminant), payload)
		runtime.KeepAlive(value)
	})
}

func NewComponentOption(value *ComponentVal) *ComponentVal {
	return newComponentVal(func(v *C.wasmtime_component_val_t) {
		var payload *C.wasmtime_component_val_t
		if value != nil {
			payload = value.ptr()
		}
		C.go_component_val_option(v, payload)
		runtime.KeepAlive(value)
	})
}

func NewComponentResult(ok bool, value *ComponentVal) *ComponentVal {
	return newComponentVal(func(v *C.wasmtime_component_val_t) {
		var payload *C.wasmtime_component_val_t
		if value != nil {
			payload = value.ptr()
		}
		C.go_component_val_result(v, C.bool(ok), payload)
		runtime.KeepAlive(value)
	})
}

func NewComponentFlags(flags []string) *ComponentVal {
	v := newComponentVal(func(v *C.wasmtime_component_val_t) { C.go_component_val_flags_init(v, C.size_t(len(flags))) })
	for i, flag := range flags {
		C.go_component_val_flags_set(&v.val, C.size_t(i), C._GoStringPtr(flag), C._GoStringLen(flag))
	}
	runtime.KeepAlive(flags)
	return v
}

func (v *ComponentVal) ptr() *C.wasmtime_component_val_t {
	if v == nil || v.closed {
		panic("component value has been closed")
	}
	return &v.val
}

func (v *ComponentVal) Kind() ComponentValKind {
	return ComponentValKind(C.go_component_val_kind(v.ptr()))
}

// Value returns the corresponding Go value. Nested component values are newly
// owned clones and must be closed independently.
func (v *ComponentVal) Value() any {
	switch v.Kind() {
	case ComponentValKindBool:
		return bool(C.go_component_val_get_bool(v.ptr()))
	case ComponentValKindS8:
		return int8(C.go_component_val_get_s8(v.ptr()))
	case ComponentValKindU8:
		return uint8(C.go_component_val_get_u8(v.ptr()))
	case ComponentValKindS16:
		return int16(C.go_component_val_get_s16(v.ptr()))
	case ComponentValKindU16:
		return uint16(C.go_component_val_get_u16(v.ptr()))
	case ComponentValKindS32:
		return int32(C.go_component_val_get_s32(v.ptr()))
	case ComponentValKindU32:
		return uint32(C.go_component_val_get_u32(v.ptr()))
	case ComponentValKindS64:
		return int64(C.go_component_val_get_s64(v.ptr()))
	case ComponentValKindU64:
		return uint64(C.go_component_val_get_u64(v.ptr()))
	case ComponentValKindF32:
		return float32(C.go_component_val_get_f32(v.ptr()))
	case ComponentValKindF64:
		return float64(C.go_component_val_get_f64(v.ptr()))
	case ComponentValKindChar:
		return rune(C.go_component_val_get_char(v.ptr()))
	case ComponentValKindString:
		return C.GoStringN(C.go_component_val_get_string(v.ptr()), C.int(C.go_component_val_get_string_len(v.ptr())))
	case ComponentValKindEnum:
		return C.GoStringN(C.go_component_val_get_enum(v.ptr()), C.int(C.go_component_val_get_enum_len(v.ptr())))
	case ComponentValKindList, ComponentValKindTuple:
		var n C.size_t
		if v.Kind() == ComponentValKindList {
			n = C.go_component_val_list_len(v.ptr())
		} else {
			n = C.go_component_val_tuple_len(v.ptr())
		}
		values := make([]*ComponentVal, int(n))
		for i := range values {
			var out C.wasmtime_component_val_t
			if v.Kind() == ComponentValKindList {
				C.go_component_val_list_get(v.ptr(), C.size_t(i), &out)
			} else {
				C.go_component_val_tuple_get(v.ptr(), C.size_t(i), &out)
			}
			values[i] = ownComponentVal(out)
		}
		return values
	case ComponentValKindRecord:
		n := int(C.go_component_val_record_len(v.ptr()))
		fields := make([]ComponentRecordField, n)
		for i := range fields {
			var out C.wasmtime_component_val_t
			fields[i].Name = C.GoStringN(C.go_component_val_record_name(v.ptr(), C.size_t(i)), C.int(C.go_component_val_record_name_len(v.ptr(), C.size_t(i))))
			C.go_component_val_record_get(v.ptr(), C.size_t(i), &out)
			fields[i].Value = ownComponentVal(out)
		}
		return fields
	case ComponentValKindVariant:
		value := ComponentVariantValue{Discriminant: C.GoStringN(C.go_component_val_variant_name(v.ptr()), C.int(C.go_component_val_variant_name_len(v.ptr())))}
		var out C.wasmtime_component_val_t
		if bool(C.go_component_val_variant_get(v.ptr(), &out)) {
			value.Value = ownComponentVal(out)
		}
		return value
	case ComponentValKindOption:
		var out C.wasmtime_component_val_t
		if !bool(C.go_component_val_option_get(v.ptr(), &out)) {
			return (*ComponentVal)(nil)
		}
		return ownComponentVal(out)
	case ComponentValKindResult:
		result := ComponentResultValue{OK: bool(C.go_component_val_result_ok(v.ptr()))}
		var out C.wasmtime_component_val_t
		if bool(C.go_component_val_result_get(v.ptr(), &out)) {
			result.Value = ownComponentVal(out)
		}
		return result
	case ComponentValKindFlags:
		n := int(C.go_component_val_flags_len(v.ptr()))
		flags := make([]string, n)
		for i := range flags {
			flags[i] = C.GoStringN(C.go_component_val_flags_get(v.ptr(), C.size_t(i)), C.int(C.go_component_val_flags_get_len(v.ptr(), C.size_t(i))))
		}
		return flags
	default:
		panic("unsupported component value kind")
	}
}

func (v *ComponentVal) Clone() *ComponentVal {
	var out C.wasmtime_component_val_t
	C.wasmtime_component_val_clone(v.ptr(), &out)
	runtime.KeepAlive(v)
	return ownComponentVal(out)
}

func (v *ComponentVal) Close() {
	if v == nil || v.closed {
		return
	}
	runtime.SetFinalizer(v, nil)
	C.wasmtime_component_val_delete(&v.val)
	v.closed = true
}

func componentValArray(size int) unsafe.Pointer {
	return C.calloc(C.size_t(size), C.size_t(C.sizeof_wasmtime_component_val_t))
}
