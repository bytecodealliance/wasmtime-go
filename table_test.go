package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTable(t *testing.T) {
	store := NewStore(NewEngine())
	ty := NewTableType(NewValType(KindFuncref), 1, true, 3)
	table, err := NewTable(store, ty, ValFuncref(nil))
	require.NoError(t, err)
	require.Equal(t, uint32(1), table.Size(store))

	f, err := table.Get(store, 0)
	require.NoError(t, err)
	require.Nil(t, f.Funcref())

	_, err = table.Get(store, 1)
	require.Error(t, err)

	err = table.Set(store, 0, ValFuncref(nil))
	require.NoError(t, err)
	err = table.Set(store, 1, ValFuncref(nil))
	require.Error(t, err)
	err = table.Set(store, 0, ValFuncref(WrapFunc(store, func() {})))
	require.NoError(t, err)
	f, err = table.Get(store, 0)
	require.NoError(t, err)
	require.NotNil(t, f.Funcref())

	prevSize, err := table.Grow(store, 1, ValFuncref(nil))
	require.NoError(t, err)
	require.Equal(t, uint32(1), prevSize)

	f, err = table.Get(store, 1)
	require.NoError(t, err)
	require.Nil(t, f.Funcref())

	called := false
	_, err = table.Grow(store, 1, ValFuncref(WrapFunc(store, func() {
		called = true
	})))
	require.NoError(t, err)
	f, err = table.Get(store, 2)
	require.NoError(t, err)
	require.False(t, called)

	_, err = f.Funcref().Call(store)
	require.NoError(t, err)
	require.True(t, called)
}
