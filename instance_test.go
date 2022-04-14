package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.NoError(t, err)
	store := NewStore(NewEngine())
	module, err := NewModule(store.Engine, wasm)
	assert.NoError(t, err)
	instance, err := NewInstance(store, module, []AsExtern{})
	assert.NoError(t, err)
	exports := instance.Exports(store)
	assert.Len(t, exports, 4)
	assert.NotNil(t, exports[0].Func())
	assert.Nil(t, exports[0].Global())
	assert.Nil(t, exports[0].Memory())
	assert.Nil(t, exports[0].Table())

	assert.Nil(t, exports[1].Func())
	assert.NotNil(t, exports[1].Global())
	assert.Nil(t, exports[1].Memory())
	assert.Nil(t, exports[1].Table())

	assert.NotNil(t, exports[2].Table())
	assert.NotNil(t, exports[3].Memory())

	f := exports[0].Func()
	g := exports[1].Global()
	table := exports[2].Table()
	m := exports[3].Memory()
	assert.Len(t, f.Type(store).Params(), 0)
	assert.Len(t, exports[0].Type(store).FuncType().Params(), 0)
	assert.Equal(t, KindI32, g.Type(store).Content().Kind())
	assert.Equal(t, KindI32, exports[1].Type(store).GlobalType().Content().Kind())
	assert.Equal(t, KindFuncref, table.Type(store).Element().Kind())
	assert.Equal(t, KindFuncref, exports[2].Type(store).TableType().Element().Kind())
	assert.Equal(t, uint64(1), m.Type(store).Minimum())
	assert.Equal(t, uint64(1), exports[3].Type(store).MemoryType().Minimum())
}

func TestInstanceBad(t *testing.T) {
	store := NewStore(NewEngine())
	wasm, err := Wat2Wasm(`(module (import "" "" (func)))`)
	assert.NoError(t, err)
	module, err := NewModule(NewEngine(), wasm)
	assert.NoError(t, err)

	// wrong number of imports
	instance, err := NewInstance(store, module, []AsExtern{})
	assert.Nil(t, instance)
	assert.Error(t, err)

	// wrong types of imports
	f := WrapFunc(store, func(a int32) {})
	instance, err = NewInstance(store, module, []AsExtern{f})
	assert.Nil(t, instance)
	assert.Error(t, err)
}

func TestInstanceGetFunc(t *testing.T) {
	wasm, err := Wat2Wasm(`
          (module
            (func (export "f") (nop))
            (global (export "g") i32 (i32.const 0))
          )
	`)
	assert.NoError(t, err)

	store := NewStore(NewEngine())
	module, err := NewModule(store.Engine, wasm)
	assert.NoError(t, err)

	instance, err := NewInstance(store, module, []AsExtern{})
	assert.NoError(t, err)

	f := instance.GetFunc(store, "f")
	assert.NotNil(t, f, "expected a function")

	_, err = f.Call(store)
	assert.NoError(t, err)

	f = instance.GetFunc(store, "g")
	assert.Nil(t, f, "expected an error")

	f = instance.GetFunc(store, "f2")
	assert.Nil(t, f, "expected an error")
}
