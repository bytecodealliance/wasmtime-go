package wasmtime

import (
	"testing"
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
	if err != nil {
		panic(err)
	}
	store := NewStore(NewEngine())
	module, err := NewModule(store.Engine, wasm)
	if err != nil {
		panic(err)
	}
	linker := NewLinker(store.Engine)
	assertNoError(linker.Define("", "f", WrapFunc(store, func() {})))
	g, err := NewGlobal(store, NewGlobalType(NewValType(KindI32), false), ValI32(0))
	assertNoError(err)
	assertNoError(linker.Define("", "g", g))
	m, err := NewMemory(store, NewMemoryType(1, true, 300))
	assertNoError(err)
	assertNoError(linker.Define("", "m", m))
	assertNoError(linker.Define("other", "m", m))

	tableWasm, err := Wat2Wasm(`(module (table (export "") 1 funcref))`)
	assertNoError(err)
	tableModule, err := NewModule(store.Engine, tableWasm)
	assertNoError(err)
	instance, err := NewInstance(store, tableModule, []AsExtern{})
	assertNoError(err)
	table := instance.Exports(store)[0].Table()
	assertNoError(linker.Define("", "t", table))

	_, err = linker.Instantiate(store, module)
	assertNoError(err)

	assertNoError(linker.DefineFunc(store, "", "", func() {}))
	assertNoError(linker.DefineInstance(store, "x", instance))
	err = linker.DefineInstance(store, "x", instance)
	if err == nil {
		panic("expected an error")
	}
}

func assertNoError(err error) {
	if err != nil {
		panic("found an error")
	}
}

func TestLinkerShadowing(t *testing.T) {
	store := NewStore(NewEngine())
	linker := NewLinker(store.Engine)
	assertNoError(linker.Define("", "f", WrapFunc(store, func() {})))
	err := linker.Define("", "f", WrapFunc(store, func() {}))
	if err == nil {
		panic("expected an error")
	}
	linker.AllowShadowing(true)
	assertNoError(linker.Define("", "f", WrapFunc(store, func() {})))
	linker.AllowShadowing(false)
	err = linker.Define("", "f", WrapFunc(store, func() {}))
	if err == nil {
		panic("expected an error")
	}
}

func TestLinkerTrap(t *testing.T) {
	store := NewStore(NewEngine())
	wasm, err := Wat2Wasm(`(func unreachable) (start 0)`)
	assertNoError(err)
	module, err := NewModule(store.Engine, wasm)
	assertNoError(err)

	linker := NewLinker(store.Engine)
	i, err := linker.Instantiate(store, module)
	if i != nil {
		panic("expected failure")
	}
	if err == nil {
		panic("expected failure")
	}
}

func TestLinkerModule(t *testing.T) {
	store := NewStore(NewEngine())
	wasm, err := Wat2Wasm(`(module
	  (func (export "f"))
	)`)
	assertNoError(err)
	module, err := NewModule(store.Engine, wasm)
	assertNoError(err)

	linker := NewLinker(store.Engine)
	err = linker.DefineModule(store, "foo", module)
	assertNoError(err)

	wasm, err = Wat2Wasm(`(module
	  (import "foo" "f" (func))
	)`)
	assertNoError(err)
	module, err = NewModule(store.Engine, wasm)
	assertNoError(err)

	_, err = linker.Instantiate(store, module)
	assertNoError(err)
}

func TestLinkerGetDefault(t *testing.T) {
	store := NewStore(NewEngine())
	linker := NewLinker(store.Engine)
	f, err := linker.GetDefault(store, "foo")
	assertNoError(err)
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
	assertNoError(err)
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
	check(err)

	wasm, err := Wat2Wasm(`
	    (module
		(import "foo" "bar" (func))
		(start 0)
	    )
	`)
	check(err)

	module, err := NewModule(engine, wasm)
	check(err)

	_, err = linker.Instantiate(NewStore(engine), module)
	check(err)
	if called != 1 {
		panic("expected a call")
	}

	_, err = linker.Instantiate(NewStore(engine), module)
	check(err)
	if called != 2 {
		panic("expected a call")
	}

	cb := func(caller *Caller, args []Val) ([]Val, *Trap) {
		called += 2
		return []Val{}, nil
	}
	ty := NewFuncType([]*ValType{}, []*ValType{})
	linker.AllowShadowing(true)
	err = linker.FuncNew("foo", "bar", ty, cb)
	check(err)

	_, err = linker.Instantiate(NewStore(engine), module)
	check(err)
	if called != 4 {
		panic("expected a call")
	}

	_, err = linker.Instantiate(NewStore(engine), module)
	check(err)
	if called != 6 {
		panic("expected a call")
	}
}
