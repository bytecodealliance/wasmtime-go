package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTableType(t *testing.T) {
	ty := NewTableType(NewValType(KindI32), 0, false, 0)
	assert.Equal(t, KindI32, ty.Element().Kind())
	assert.Equal(t, uint32(0), ty.Minimum())

	present, _ := ty.Maximum()
	assert.False(t, present)

	ty = NewTableType(NewValType(KindF64), 1, true, 129)
	assert.Equal(t, KindF64, ty.Element().Kind())
	assert.Equal(t, uint32(1), ty.Minimum())

	present, max := ty.Maximum()
	assert.True(t, present)
	assert.Equal(t, uint32(129), max)

	ty2 := ty.AsExternType().TableType()
	assert.NotNil(t, ty2)
	assert.Nil(t, ty.AsExternType().FuncType())
	assert.Nil(t, ty.AsExternType().GlobalType())
	assert.Nil(t, ty.AsExternType().MemoryType())
}
