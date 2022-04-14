package wasmtime

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalType(t *testing.T) {
	ty := NewGlobalType(NewValType(KindI32), true)
	assert.Equal(t, KindI32, ty.Content().Kind())
	assert.True(t, ty.Mutable())

	content := ty.Content()
	runtime.GC()
	assert.Equal(t, KindI32, content.Kind())

	ty = NewGlobalType(NewValType(KindI32), true)
	ty2 := ty.AsExternType().GlobalType()
	assert.NotNil(t, ty2)
	assert.Nil(t, ty.AsExternType().FuncType())
	assert.Nil(t, ty.AsExternType().MemoryType())
	assert.Nil(t, ty.AsExternType().TableType())
}
