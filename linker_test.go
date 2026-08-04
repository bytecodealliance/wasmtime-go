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
	mt, err := NewMemoryType(1, true, 300, false)
	require.NoError(t, err)
	m, err := NewMemory(store, mt)
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
	var callers []string
	err := linker.FuncWrap("foo", "bar", func(c *Caller) {
		called += 1

		data := c.Data()
		if data != nil {
			callers = append(callers, data.(string))
		}
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

	_, err = linker.Instantiate(NewStoreWithData(engine, "one"), module)
	require.NoError(t, err)
	require.Equal(t, 1, called, "expected a call")
	require.Equal(t, []string{"one"}, callers)

	_, err = linker.Instantiate(NewStore(engine), module)
	require.NoError(t, err)
	require.Equal(t, 2, called, "expected a call")

	cb := func(caller *Caller, args []Val) ([]Val, *Trap) {
		called += 2
		data := caller.Data()
		if data != nil {
			callers = append(callers, data.(string))
		}

		return []Val{}, nil
	}
	ty := NewFuncType([]*ValType{}, []*ValType{})
	linker.AllowShadowing(true)
	err = linker.FuncNew("foo", "bar", ty, cb)
	require.NoError(t, err)

	_, err = linker.Instantiate(NewStoreWithData(engine, "two"), module)
	require.NoError(t, err)
	require.Equal(t, 4, called, "expected a call")
	require.Equal(t, []string{"one", "two"}, callers)

	_, err = linker.Instantiate(NewStore(engine), module)
	require.NoError(t, err)
	require.Equal(t, 6, called, "expected a call")
}

func TestLinkerFuncWrapRejectsInvalidUTF8WithoutRetainingCallback(t *testing.T) {
	engine := NewEngine()
	linker := NewLinker(engine)
	defer linker.Close()

	for _, tc := range []struct {
		module string
		name   string
	}{
		{module: string([]byte{0xff}), name: "host"},
		{module: "host", name: string([]byte{0xff})},
	} {
		_, before := engineFuncCounts()
		err := linker.FuncWrap(tc.module, tc.name, func() {})
		require.EqualError(t, err, "module and name must be valid UTF-8")
		_, after := engineFuncCounts()
		require.Equal(t, before, after)
	}
}

func TestLinkerFuncNewRejectsInvalidUTF8WithoutRetainingCallback(t *testing.T) {
	engine := NewEngine()
	linker := NewLinker(engine)
	defer linker.Close()
	ty := NewFuncType(nil, nil)
	defer ty.Close()

	for _, tc := range []struct {
		module string
		name   string
	}{
		{module: string([]byte{0xff}), name: "host"},
		{module: "host", name: string([]byte{0xff})},
	} {
		before, _ := engineFuncCounts()
		err := linker.FuncNew(tc.module, tc.name, ty, func(*Caller, []Val) ([]Val, *Trap) {
			return nil, nil
		})
		require.EqualError(t, err, "module and name must be valid UTF-8")
		after, _ := engineFuncCounts()
		require.Equal(t, before, after)
	}
}

func engineFuncCounts() (funcNew, funcWrap int) {
	gEngineFuncLock.Lock()
	defer gEngineFuncLock.Unlock()
	return len(gEngineFuncNew), len(gEngineFuncWrap)
}

func TestLinkerFuncsDoNotRetainCallbacksWhenLinkerIsClosed(t *testing.T) {
	engine := NewEngine()
	linker := NewLinker(engine)
	linker.Close()
	ty := NewFuncType(nil, nil)
	defer ty.Close()

	beforeNew, beforeWrap := engineFuncCounts()
	require.PanicsWithValue(t, "object has been closed already", func() {
		linker.FuncWrap("host", "wrap", func() {})
	})
	require.PanicsWithValue(t, "object has been closed already", func() {
		linker.FuncNew("host", "new", ty, func(*Caller, []Val) ([]Val, *Trap) {
			return nil, nil
		})
	})
	afterNew, afterWrap := engineFuncCounts()
	require.Equal(t, beforeNew, afterNew)
	require.Equal(t, beforeWrap, afterWrap)
}

func TestLinkerDefineUnknownImportsAsTraps(t *testing.T) {
	engine := NewEngine()
	wasm, err := Wat2Wasm(`
	    (module
		(import "unknown" "f" (func))
		(func (export "run") call 0)
	    )
	`)
	require.NoError(t, err)
	module, err := NewModule(engine, wasm)
	require.NoError(t, err)

	store := NewStore(engine)
	linker := NewLinker(engine)
	defer linker.Close()
	_, err = linker.Instantiate(store, module)
	require.Error(t, err)

	err = linker.DefineUnknownImportsAsTraps(module)
	require.NoError(t, err)
	instance, err := linker.Instantiate(store, module)
	require.NoError(t, err)

	run := instance.GetFunc(store, "run")
	require.NotNil(t, run)
	_, err = run.Call(store)
	require.Error(t, err)
}

func TestLinkerDefineUnknownImportsAsDefaultValues(t *testing.T) {
	engine := NewEngine()
	wasm, err := Wat2Wasm(`
	    (module
		(import "unknown" "f" (func (result i32)))
		(func (export "run") (result i32) call 0)
	    )
	`)
	require.NoError(t, err)
	module, err := NewModule(engine, wasm)
	require.NoError(t, err)

	store := NewStore(engine)
	linker := NewLinker(engine)
	defer linker.Close()

	_, err = linker.Instantiate(store, module)
	require.Error(t, err)

	err = linker.DefineUnknownImportsAsDefaultValues(store, module)
	require.NoError(t, err)
	instance, err := linker.Instantiate(store, module)
	require.NoError(t, err)

	run := instance.GetFunc(store, "run")
	require.NotNil(t, run)
	result, err := run.Call(store)
	require.NoError(t, err)
	require.Equal(t, int32(0), result)
}

func TestPanicInHostFunctionDuringInstantiate(t *testing.T) {
	engine := NewEngine()
	linker := NewLinker(engine)
	store := NewStore(engine)
	err := linker.FuncWrap("foo", "bar", func() {
		panic("hi")
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
	require.PanicsWithValue(t, "hi", func() {
		linker.Instantiate(store, module)
	})
}

func TestPanicInHostFunctionDuringCall(t *testing.T) {
	engine := NewEngine()
	linker := NewLinker(engine)
	store := NewStore(engine)
	err := linker.FuncWrap("foo", "bar", func() {
		panic("hi")
	})
	require.NoError(t, err)

	wasm, err := Wat2Wasm(`
	    (module
		(import "foo" "bar" (func))
		(func (export "foo") call 0)
	    )
	`)
	require.NoError(t, err)

	module, err := NewModule(engine, wasm)
	require.NoError(t, err)
	instance, err := linker.Instantiate(store, module)
	require.NoError(t, err)
	foo := instance.GetFunc(store, "foo")
	require.NotNil(t, foo)
	require.PanicsWithValue(t, "hi", func() {
		foo.Call(store)
	})
}

func TestPanicInHostFunctionDuringCallWithCaller(t *testing.T) {
	engine := NewEngine()
	linker := NewLinker(engine)
	store := NewStore(engine)
	err := linker.FuncWrap("foo", "bar", func(c *Caller) int32 {
		panic("hi")
	})
	require.NoError(t, err)

	wasm, err := Wat2Wasm(`
	    (module
		(import "foo" "bar" (func (result i32)))
		(func (export "foo") call 0 drop)
	    )
	`)
	require.NoError(t, err)

	module, err := NewModule(engine, wasm)
	require.NoError(t, err)
	instance, err := linker.Instantiate(store, module)
	require.NoError(t, err)
	foo := instance.GetFunc(store, "foo")
	require.NotNil(t, foo)
	require.PanicsWithValue(t, "hi", func() {
		foo.Call(store)
	})
}
