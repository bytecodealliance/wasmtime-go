package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImportType(t *testing.T) {
	fty := NewFuncType([]*ValType{}, []*ValType{})
	ty := NewImportType("a", "b", fty)
	assert.Equal(t, "a", ty.Module())
	assert.Equal(t, "b", *ty.Name())
	assert.NotNil(t, ty.Type().FuncType())

	gty := NewGlobalType(NewValType(KindI32), true)
	ty = NewImportType("", "", gty.AsExternType())
	assert.Empty(t, ty.Module())
	assert.Empty(t, *ty.Name())
	assert.NotNil(t, ty.Type().GlobalType())
}
