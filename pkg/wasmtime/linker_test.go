package wasmtime

import "testing"

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
	module, err := NewModule(store, wasm)
	if err != nil {
		panic(err)
	}
	linker := NewLinker(module.Store)
	assertNoError(linker.Define("", "f", WrapFunc(store, func() {})))
	g, err := NewGlobal(store, NewGlobalType(NewValType(KindI32), false), ValI32(0))
	assertNoError(err)
	assertNoError(linker.Define("", "g", g))
	m := NewMemory(store, NewMemoryType(Limits{Min: 1, Max: 0xffffffff}))
	assertNoError(linker.Define("", "m", m))
	assertNoError(linker.Define("other", "m", m.AsExtern()))

	table_wasm, err := Wat2Wasm(`(module (table (export "") 1 funcref))`)
	assertNoError(err)
	table_module, err := NewModule(store, table_wasm)
	assertNoError(err)
	instance, err := NewInstance(table_module, []*Extern{})
	assertNoError(err)
	table := instance.Exports()[0].Table()
	assertNoError(linker.Define("", "t", table))

	_, err = linker.Instantiate(module)
	assertNoError(err)

	assertNoError(linker.DefineFunc("", "", func() {}))
	assertNoError(linker.DefineInstance("x", instance))
	err = linker.DefineInstance("x", instance)
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
	linker := NewLinker(store)
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
	module, err := NewModule(store, wasm)
	assertNoError(err)

	linker := NewLinker(store)
	i, err := linker.Instantiate(module)
	if i != nil {
		panic("expected failure")
	}
	if err == nil {
		panic("expected failure")
	}
}
