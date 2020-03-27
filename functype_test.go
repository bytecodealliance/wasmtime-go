package wasmtime

import "testing"

func TestFuncType(t *testing.T) {
	NewFuncType(make([]*ValType, 0), make([]*ValType, 0))

	i32 := NewValType(KindI32)
	i64 := NewValType(KindI64)
	NewFuncType([]*ValType{i32}, []*ValType{})
	NewFuncType([]*ValType{i32}, []*ValType{i32})
	NewFuncType([]*ValType{}, []*ValType{i32})
	NewFuncType([]*ValType{i32, i64, i64}, []*ValType{i32, i64, i64})

	ty := NewFuncType([]*ValType{}, []*ValType{})
	if len(ty.Params()) != 0 {
		panic("expect 0 params")
	}
	if len(ty.Results()) != 0 {
		panic("expect 0 results")
	}

	ty = NewFuncType([]*ValType{i32, i64, i64}, []*ValType{i32, i64, i64})

	params := ty.Params()
	if len(params) != 3 {
		panic("expect 3 params")
	}
	if params[0].Kind() != KindI32 {
		panic("unexpected kind")
	}
	if params[1].Kind() != KindI64 {
		panic("unexpected kind")
	}
	if params[2].Kind() != KindI64 {
		panic("unexpected kind")
	}
	results := ty.Results()
	if len(results) != 3 {
		panic("expect 3 results")
	}
	if results[0].Kind() != KindI32 {
		panic("unexpected kind")
	}
	if results[1].Kind() != KindI64 {
		panic("unexpected kind")
	}
	if results[2].Kind() != KindI64 {
		panic("unexpected kind")
	}
}
