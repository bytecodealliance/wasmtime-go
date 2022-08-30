package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstance(t *testing.T) {
	wasm, err := Wat2Wasm(`
          (module
            (func (export "f"))
            (global (export "g") i32 (i32.const 0))
            (table (export "t") 1 funcref)
            (memory (export "m") 1)
          )
        `)
	require.NoError(t, err)
	store := NewStore(NewEngine())
	module, err := NewModule(store.Engine, wasm)
	require.NoError(t, err)
	instance, err := NewInstance(store, module, []AsExtern{})
	require.NoError(t, err)
	exports := instance.Exports(store)
	require.Len(t, exports, 4)
	require.NotNil(t, exports[0].Func())
	require.Nil(t, exports[0].Global())
	require.Nil(t, exports[0].Memory())
	require.Nil(t, exports[0].Table())

	require.Nil(t, exports[1].Func())
	require.NotNil(t, exports[1].Global())
	require.Nil(t, exports[1].Memory())
	require.Nil(t, exports[1].Table())

	require.NotNil(t, exports[2].Table())
	require.NotNil(t, exports[3].Memory())

	f := exports[0].Func()
	g := exports[1].Global()
	table := exports[2].Table()
	m := exports[3].Memory()
	require.Len(t, f.Type(store).Params(), 0)
	require.Len(t, exports[0].Type(store).FuncType().Params(), 0)
	require.Equal(t, KindI32, g.Type(store).Content().Kind())
	require.Equal(t, KindI32, exports[1].Type(store).GlobalType().Content().Kind())
	require.Equal(t, KindFuncref, table.Type(store).Element().Kind())
	require.Equal(t, KindFuncref, exports[2].Type(store).TableType().Element().Kind())
	require.Equal(t, uint64(1), m.Type(store).Minimum())
	require.Equal(t, uint64(1), exports[3].Type(store).MemoryType().Minimum())
}

func TestInstanceBad(t *testing.T) {
	store := NewStore(NewEngine())
	wasm, err := Wat2Wasm(`(module (import "" "" (func)))`)
	require.NoError(t, err)
	module, err := NewModule(NewEngine(), wasm)
	require.NoError(t, err)

	// wrong number of imports
	instance, err := NewInstance(store, module, []AsExtern{})
	require.Nil(t, instance)
	require.Error(t, err)

	// wrong types of imports
	f := WrapFunc(store, func(a int32) {})
	instance, err = NewInstance(store, module, []AsExtern{f})
	require.Nil(t, instance)
	require.Error(t, err)
}

func TestInstanceGetFunc(t *testing.T) {
	wasm, err := Wat2Wasm(`
          (module
            (func (export "f") (nop))
            (global (export "g") i32 (i32.const 0))
          )
	`)
	require.NoError(t, err)

	store := NewStore(NewEngine())
	module, err := NewModule(store.Engine, wasm)
	require.NoError(t, err)

	instance, err := NewInstance(store, module, []AsExtern{})
	require.NoError(t, err)

	f := instance.GetFunc(store, "f")
	require.NotNil(t, f, "expected a function")

	_, err = f.Call(store)
	require.NoError(t, err)

	f = instance.GetFunc(store, "g")
	require.Nil(t, f, "expected an error")

	f = instance.GetFunc(store, "f2")
	require.Nil(t, f, "expected an error")
}
