package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImportType(t *testing.T) {
	fty := NewFuncType([]*ValType{}, []*ValType{})
	ty := NewImportType("a", "b", fty)
	require.Equal(t, "a", ty.Module())
	require.Equal(t, "b", *ty.Name())
	require.NotNil(t, ty.Type().FuncType())

	gty := NewGlobalType(NewValType(KindI32), true)
	ty = NewImportType("", "", gty.AsExternType())
	require.Empty(t, ty.Module())
	require.Empty(t, *ty.Name())
	require.NotNil(t, ty.Type().GlobalType())
}
