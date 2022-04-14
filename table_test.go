package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTable(t *testing.T) {
	store := NewStore(NewEngine())
	ty := NewTableType(NewValType(KindFuncref), 1, true, 3)
	table, err := NewTable(store, ty, ValFuncref(nil))
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), table.Size(store))

	f, err := table.Get(store, 0)
	assert.NoError(t, err)
	assert.Nil(t, f.Funcref())

	f, err = table.Get(store, 1)
	assert.Error(t, err)

	err = table.Set(store, 0, ValFuncref(nil))
	assert.NoError(t, err)
	err = table.Set(store, 1, ValFuncref(nil))
	assert.Error(t, err)
	err = table.Set(store, 0, ValFuncref(WrapFunc(store, func() {})))
	assert.NoError(t, err)
	f, err = table.Get(store, 0)
	assert.NoError(t, err)
	assert.NotNil(t, f.Funcref())

	prevSize, err := table.Grow(store, 1, ValFuncref(nil))
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), prevSize)

	f, err = table.Get(store, 1)
	assert.NoError(t, err)
	assert.Nil(t, f.Funcref())

	called := false
	_, err = table.Grow(store, 1, ValFuncref(WrapFunc(store, func() {
		called = true
	})))
	assert.NoError(t, err)
	f, err = table.Get(store, 2)
	assert.NoError(t, err)
	assert.False(t, called)

	_, err = f.Funcref().Call(store)
	assert.NoError(t, err)
	assert.True(t, called)
}
