package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinker(t *testing.T) {
	wasm, err := Wat2Wasm(`
          (module
	    (import "" "f" (func))
	    (import "" "g" (global i32))
	    (import "" "t" (table 1 funcref))
	    (import "" "m" (memory 1))
          )
        `)
	assert.NoError(t, err)
	store := NewStore(NewEngine())
	module, err := NewModule(store.Engine, wasm)
	assert.NoError(t, err)
	linker := NewLinker(store.Engine)
	assert.NoError(t, linker.Define("", "f", WrapFunc(store, func() {})))
	g, err := NewGlobal(store, NewGlobalType(NewValType(KindI32), false), ValI32(0))
	assert.NoError(t, err)
	assert.NoError(t, linker.Define("", "g", g))
	m, err := NewMemory(store, NewMemoryType(1, true, 300))
	assert.NoError(t, err)
	assert.NoError(t, linker.Define("", "m", m))
	assert.NoError(t, linker.Define("other", "m", m))

	tableWasm, err := Wat2Wasm(`(module (table (export "") 1 funcref))`)
	assert.NoError(t, err)
	tableModule, err := NewModule(store.Engine, tableWasm)
	assert.NoError(t, err)
	instance, err := NewInstance(store, tableModule, []AsExtern{})
	assert.NoError(t, err)
	table := instance.Exports(store)[0].Table()
	assert.NoError(t, linker.Define("", "t", table))

	_, err = linker.Instantiate(store, module)
	assert.NoError(t, err)

	assert.NoError(t, linker.DefineFunc(store, "", "", func() {}))
	assert.NoError(t, linker.DefineInstance(store, "x", instance))
	err = linker.DefineInstance(store, "x", instance)
	assert.Error(t, err)
}

func TestLinkerShadowing(t *testing.T) {
	store := NewStore(NewEngine())
	linker := NewLinker(store.Engine)
	assert.NoError(t, linker.Define("", "f", WrapFunc(store, func() {})))
	err := linker.Define("", "f", WrapFunc(store, func() {}))
	assert.NoError(t, err)

	linker.AllowShadowing(true)
	assert.NoError(t, linker.Define("", "f", WrapFunc(store, func() {})))
	linker.AllowShadowing(false)
	err = linker.Define("", "f", WrapFunc(store, func() {}))
	assert.NoError(t, err)
}

func TestLinkerTrap(t *testing.T) {
	store := NewStore(NewEngine())
	wasm, err := Wat2Wasm(`(func unreachable) (start 0)`)
	assert.NoError(t, err)
	module, err := NewModule(store.Engine, wasm)
	assert.NoError(t, err)

	linker := NewLinker(store.Engine)
	inst, err := linker.Instantiate(store, module)
	assert.Nil(t, inst)
	assert.Error(t, err)
}

func TestLinkerModule(t *testing.T) {
	store := NewStore(NewEngine())
	wasm, err := Wat2Wasm(`(module
	  (func (export "f"))
	)`)
	assert.NoError(t, err)
	module, err := NewModule(store.Engine, wasm)
	assert.NoError(t, err)

	linker := NewLinker(store.Engine)
	err = linker.DefineModule(store, "foo", module)
	assert.NoError(t, err)

	wasm, err = Wat2Wasm(`(module
	  (import "foo" "f" (func))
	)`)
	assert.NoError(t, err)
	module, err = NewModule(store.Engine, wasm)
	assert.NoError(t, err)

	_, err = linker.Instantiate(store, module)
	assert.NoError(t, err)
}

func TestLinkerGetDefault(t *testing.T) {
	store := NewStore(NewEngine())
	linker := NewLinker(store.Engine)
	f, err := linker.GetDefault(store, "foo")
	assert.NoError(t, err)
	f.Call(store)
}

func TestLinkerGetOneByName(t *testing.T) {
	store := NewStore(NewEngine())
	linker := NewLinker(store.Engine)
	f := linker.Get(store, "foo", "bar")
	if f != nil {
		panic("expected nil")
	}

	err := linker.DefineFunc(store, "foo", "baz", func() {})
	assert.NoError(t, err)
	f = linker.Get(store, "foo", "baz")
	f.Func().Call(store)
}

func TestLinkerFuncs(t *testing.T) {
	engine := NewEngine()
	linker := NewLinker(engine)
	var called int
	err := linker.FuncWrap("foo", "bar", func() {
		called += 1
	})
	assert.NoError(t, err)

	wasm, err := Wat2Wasm(`
	    (module
		(import "foo" "bar" (func))
		(start 0)
	    )
	`)
	assert.NoError(t, err)

	module, err := NewModule(engine, wasm)
	assert.NoError(t, err)

	_, err = linker.Instantiate(NewStore(engine), module)
	assert.NoError(t, err)
	assert.Equal(t, 1, called, "expected a call")

	_, err = linker.Instantiate(NewStore(engine), module)
	assert.NoError(t, err)
	assert.Equal(t, 2, called, "expected a call")

	cb := func(caller *Caller, args []Val) ([]Val, *Trap) {
		called += 2
		return []Val{}, nil
	}
	ty := NewFuncType([]*ValType{}, []*ValType{})
	linker.AllowShadowing(true)
	err = linker.FuncNew("foo", "bar", ty, cb)
	assert.NoError(t, err)

	_, err = linker.Instantiate(NewStore(engine), module)
	assert.NoError(t, err)
	assert.Equal(t, 4, called, "expected a call")

	_, err = linker.Instantiate(NewStore(engine), module)
	assert.NoError(t, err)
	assert.Equal(t, 6, called, "expected a call")
}
