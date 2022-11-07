package v2

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGlobalType(t *testing.T) {
	ty := NewGlobalType(NewValType(KindI32), true)
	require.Equal(t, KindI32, ty.Content().Kind())
	require.True(t, ty.Mutable())

	content := ty.Content()
	runtime.GC()
	require.Equal(t, KindI32, content.Kind())

	ty = NewGlobalType(NewValType(KindI32), true)
	ty2 := ty.AsExternType().GlobalType()
	require.NotNil(t, ty2)
	require.Nil(t, ty.AsExternType().FuncType())
	require.Nil(t, ty.AsExternType().MemoryType())
	require.Nil(t, ty.AsExternType().TableType())
}
