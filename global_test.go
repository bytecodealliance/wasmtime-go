package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobal(t *testing.T) {
	store := NewStore(NewEngine())
	g, err := NewGlobal(store, NewGlobalType(NewValType(KindI32), true), ValI32(100))
	assert.NoError(t, err)
	assert.Equal(t, int32(100), g.Get(store).I32())

	g.Set(store, ValI32(200))
	assert.Equal(t, int32(200), g.Get(store).I32())

	_, err = NewGlobal(store, NewGlobalType(NewValType(KindI64), true), ValI32(100))
	assert.Error(t, err, "should fail to create global")
	err = g.Set(store, ValI64(200))
	assert.Error(t, err, "should fail to set global")
}
