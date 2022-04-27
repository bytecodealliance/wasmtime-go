package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFuncType(t *testing.T) {
	NewFuncType(make([]*ValType, 0), make([]*ValType, 0))

	i32 := NewValType(KindI32)
	i64 := NewValType(KindI64)
	NewFuncType([]*ValType{i32}, []*ValType{})
	NewFuncType([]*ValType{i32}, []*ValType{i32})
	NewFuncType([]*ValType{}, []*ValType{i32})
	NewFuncType([]*ValType{i32, i64, i64}, []*ValType{i32, i64, i64})

	ty := NewFuncType([]*ValType{}, []*ValType{})
	require.Len(t, ty.Params(), 0)
	require.Len(t, ty.Results(), 0)

	ty = NewFuncType([]*ValType{i32, i64, i64}, []*ValType{i32, i64, i64})

	params := ty.Params()
	require.Len(t, ty.Params(), 3)
	require.Equal(t, KindI32, params[0].Kind())
	require.Equal(t, KindI64, params[1].Kind())
	require.Equal(t, KindI64, params[2].Kind())

	results := ty.Results()
	require.Len(t, ty.Params(), 3)
	require.Equal(t, KindI32, results[0].Kind())
	require.Equal(t, KindI64, results[1].Kind())
	require.Equal(t, KindI64, results[2].Kind())

	ty = NewFuncType([]*ValType{}, []*ValType{})
	ty2 := ty.AsExternType().FuncType()
	require.NotNil(t, ty2)
	require.Nil(t, ty.AsExternType().GlobalType())
	require.Nil(t, ty.AsExternType().MemoryType())
	require.Nil(t, ty.AsExternType().TableType())
}
