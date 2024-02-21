package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	store := NewStore(NewEngine())
	module, err := NewModule(store.Engine, wasm)
	require.NoError(t, err)
	linker := NewLinker(store.Engine)
	defer linker.Close()
	require.NoError(t, linker.Define(store, "", "f", WrapFunc(store, func() {})))
	g, err := NewGlobal(store, NewGlobalType(NewValType(KindI32), false), ValI32(0))
	require.NoError(t, err)
	require.NoError(t, linker.Define(store, "", "g", g))
	m, err := NewMemory(store, NewMemoryType(1, true, 300))
	require.NoError(t, err)
	require.NoError(t, linker.Define(store, "", "m", m))
	require.NoError(t, linker.Define(store, "other", "m", m))

	tableWasm, err := Wat2Wasm(`(module (table (export "") 1 funcref))`)
	require.NoError(t, err)
	tableModule, err := NewModule(store.Engine, tableWasm)
	require.NoError(t, err)
	instance, err := NewInstance(store, tableModule, []AsExtern{})
	require.NoError(t, err)
	table := instance.Exports(store)[0].Table()
	require.NoError(t, linker.Define(store, "", "t", table))

	_, err = linker.Instantiate(store, module)
	require.NoError(t, err)

	require.NoError(t, linker.DefineFunc(store, "", "", func() {}))
	require.NoError(t, linker.DefineInstance(store, "x", instance))
	err = linker.DefineInstance(store, "x", instance)
	require.Error(t, err)
}

func TestLinkerShadowing(t *testing.T) {
	store := NewStore(NewEngine())
	linker := NewLinker(store.Engine)
	require.NoError(t, linker.Define(store, "", "f", WrapFunc(store, func() {})))
	err := linker.Define(store, "", "f", WrapFunc(store, func() {}))
	require.Error(t, err)

	linker.AllowShadowing(true)
	require.NoError(t, linker.Define(store, "", "f", WrapFunc(store, func() {})))
	linker.AllowShadowing(false)
	err = linker.Define(store, "", "f", WrapFunc(store, func() {}))
	require.Error(t, err)
}

func TestLinkerTrap(t *testing.T) {
	store := NewStore(NewEngine())
	wasm, err := Wat2Wasm(`(func unreachable) (start 0)`)
	require.NoError(t, err)
	module, err := NewModule(store.Engine, wasm)
	require.NoError(t, err)

	linker := NewLinker(store.Engine)
	inst, err := linker.Instantiate(store, module)
	require.Nil(t, inst)
	require.Error(t, err)
}

func TestLinkerModule(t *testing.T) {
	store := NewStore(NewEngine())
	wasm, err := Wat2Wasm(`(module
	  (func (export "f"))
	)`)
	require.NoError(t, err)
	module, err := NewModule(store.Engine, wasm)
	require.NoError(t, err)

	linker := NewLinker(store.Engine)
	err = linker.DefineModule(store, "foo", module)
	require.NoError(t, err)

	wasm, err = Wat2Wasm(`(module
	  (import "foo" "f" (func))
	)`)
	require.NoError(t, err)
	module, err = NewModule(store.Engine, wasm)
	require.NoError(t, err)

	_, err = linker.Instantiate(store, module)
	require.NoError(t, err)
}

func TestLinkerGetDefault(t *testing.T) {
	store := NewStore(NewEngine())
	linker := NewLinker(store.Engine)
	f, err := linker.GetDefault(store, "foo")
	require.NoError(t, err)
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
	require.NoError(t, err)
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
	require.NoError(t, err)

	wasm, err := Wat2Wasm(`
	    (module
		(import "foo" "bar" (func))
		(start 0)
	    )
	`)
	require.NoError(t, err)

	module, err := NewModule(engine, wasm)
	require.NoError(t, err)

	_, err = linker.Instantiate(NewStore(engine), module)
	require.NoError(t, err)
	require.Equal(t, 1, called, "expected a call")

	_, err = linker.Instantiate(NewStore(engine), module)
	require.NoError(t, err)
	require.Equal(t, 2, called, "expected a call")

	cb := func(caller *Caller, args []Val) ([]Val, *Trap) {
		called += 2
		return []Val{}, nil
	}
	ty := NewFuncType([]*ValType{}, []*ValType{})
	linker.AllowShadowing(true)
	err = linker.FuncNew("foo", "bar", ty, cb)
	require.NoError(t, err)

	_, err = linker.Instantiate(NewStore(engine), module)
	require.NoError(t, err)
	require.Equal(t, 4, called, "expected a call")

	_, err = linker.Instantiate(NewStore(engine), module)
	require.NoError(t, err)
	require.Equal(t, 6, called, "expected a call")
}
