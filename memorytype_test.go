package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryType(t *testing.T) {
	ty := NewMemoryType(0, true, 100)
	ty.Minimum()
	ty.Maximum()

	ty2 := ty.AsExternType().MemoryType()
	assert.NotNil(t, ty2)
	assert.Nil(t, ty.AsExternType().FuncType())
	assert.Nil(t, ty.AsExternType().GlobalType())
	assert.Nil(t, ty.AsExternType().TableType())
}

func TestMemoryType64(t *testing.T) {
	ty := NewMemoryType64(0x100000000, true, 0x100000001)
	assert.Equal(t, uint64(0x100000000), ty.Minimum())

	present, max := ty.Maximum()
	assert.Equal(t, uint64(0x100000001), max)
	assert.True(t, present)
}
