package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTableType(t *testing.T) {
	ty := NewTableType(NewValType(KindI32), 0, false, 0)
	require.Equal(t, KindI32, ty.Element().Kind())
	require.Equal(t, uint32(0), ty.Minimum())

	present, _ := ty.Maximum()
	require.False(t, present)

	ty = NewTableType(NewValType(KindF64), 1, true, 129)
	require.Equal(t, KindF64, ty.Element().Kind())
	require.Equal(t, uint32(1), ty.Minimum())

	present, max := ty.Maximum()
	require.True(t, present)
	require.Equal(t, uint32(129), max)

	ty2 := ty.AsExternType().TableType()
	require.NotNil(t, ty2)
	require.Nil(t, ty.AsExternType().FuncType())
	require.Nil(t, ty.AsExternType().GlobalType())
	require.Nil(t, ty.AsExternType().MemoryType())
}
