package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemoryType(t *testing.T) {
	ty, err := NewMemoryType(0, true, 100, false)
	require.NoError(t, err)
	ty.Minimum()
	ty.Maximum()

	ty2 := ty.AsExternType().MemoryType()
	require.NotNil(t, ty2)
	require.Nil(t, ty.AsExternType().FuncType())
	require.Nil(t, ty.AsExternType().GlobalType())
	require.Nil(t, ty.AsExternType().TableType())
}

func TestMemoryType64(t *testing.T) {
	ty, err := NewMemoryType64(0x100000000, true, 0x100000001, false)
	require.NoError(t, err)
	require.Equal(t, uint64(0x100000000), ty.Minimum())
	require.True(t, ty.Is64())
	require.False(t, ty.IsShared())

	present, max := ty.Maximum()
	require.Equal(t, uint64(0x100000001), max)
	require.True(t, present)
}
