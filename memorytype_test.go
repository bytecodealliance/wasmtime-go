package v2

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemoryType(t *testing.T) {
	ty := NewMemoryType(0, true, 100)
	ty.Minimum()
	ty.Maximum()

	ty2 := ty.AsExternType().MemoryType()
	require.NotNil(t, ty2)
	require.Nil(t, ty.AsExternType().FuncType())
	require.Nil(t, ty.AsExternType().GlobalType())
	require.Nil(t, ty.AsExternType().TableType())
}

func TestMemoryType64(t *testing.T) {
	ty := NewMemoryType64(0x100000000, true, 0x100000001)
	require.Equal(t, uint64(0x100000000), ty.Minimum())

	present, max := ty.Maximum()
	require.Equal(t, uint64(0x100000001), max)
	require.True(t, present)
}
