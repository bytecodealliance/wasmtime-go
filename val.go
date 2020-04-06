package wasmtime

// #include <wasm.h>
//
// void go_init_i32(wasm_val_t *val, int32_t i) { val->of.i32 = i; }
// void go_init_i64(wasm_val_t *val, int64_t i) { val->of.i64 = i; }
// void go_init_f32(wasm_val_t *val, float i) { val->of.f32 = i; }
// void go_init_f64(wasm_val_t *val, double i) { val->of.f64 = i; }
//
// int32_t go_get_i32(wasm_val_t *val) { return val->of.i32; }
// int64_t go_get_i64(wasm_val_t *val) { return val->of.i64; }
// float go_get_f32(wasm_val_t *val) { return val->of.f32; }
// double go_get_f64(wasm_val_t *val) { return val->of.f64; }
import "C"

type Val struct {
	raw C.wasm_val_t
}

func ValI32(val int32) Val {
	ret := Val{raw: C.wasm_val_t{kind: C.WASM_I32}}
	C.go_init_i32(&ret.raw, C.int32_t(val))
	return ret
}

func ValI64(val int64) Val {
	ret := Val{raw: C.wasm_val_t{kind: C.WASM_I64}}
	C.go_init_i64(&ret.raw, C.int64_t(val))
	return ret
}

func ValF32(val float32) Val {
	ret := Val{raw: C.wasm_val_t{kind: C.WASM_F32}}
	C.go_init_f32(&ret.raw, C.float(val))
	return ret
}

func ValF64(val float64) Val {
	ret := Val{raw: C.wasm_val_t{kind: C.WASM_F64}}
	C.go_init_f64(&ret.raw, C.double(val))
	return ret
}

// Returns the kind of value that this `Val` contains.
func (v Val) Kind() ValKind {
	return ValKind(v.raw.kind)
}

// Returns the underlying 32-bit integer if this is an `i32`, or panics.
func (v Val) I32() int32 {
	if v.Kind() != KindI32 {
		panic("not an i32")
	}
	return int32(C.go_get_i32(&v.raw))
}

// Returns the underlying 64-bit integer if this is an `i64`, or panics.
func (v Val) I64() int64 {
	if v.Kind() != KindI64 {
		panic("not an i64")
	}
	return int64(C.go_get_i64(&v.raw))
}

// Returns the underlying 32-bit float if this is an `f32`, or panics.
func (v Val) F32() float32 {
	if v.Kind() != KindF32 {
		panic("not an f32")
	}
	return float32(C.go_get_f32(&v.raw))
}

// Returns the underlying 64-bit float if this is an `f64`, or panics.
func (v Val) F64() float64 {
	if v.Kind() != KindF64 {
		panic("not an f64")
	}
	return float64(C.go_get_f64(&v.raw))
}

// Returns the underlying 64-bit float if this is an `f64`, or panics.
func (v Val) Get() interface{} {
	switch v.Kind() {
	case KindI32:
		return v.I32()
	case KindI64:
		return v.I64()
	case KindF32:
		return v.F32()
	case KindF64:
		return v.F64()
	}
	panic("failed to get value of `Val`")
}
