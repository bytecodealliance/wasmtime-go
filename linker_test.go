package wasmtime

import (
	"fmt"
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
	linker := NewLinker(store)
	assertNoError(linker.Define("", "f", WrapFunc(store, func() {})))
	g, err := NewGlobal(store, NewGlobalType(NewValType(KindI32), false), ValI32(0))
	assertNoError(err)
	assertNoError(linker.Define("", "g", g))
	m := NewMemory(store, NewMemoryType(Limits{Min: 1, Max: 0xffffffff}))
	assertNoError(linker.Define("", "m", m))
	assertNoError(linker.Define("other", "m", m.AsExtern()))

	tableWasm, err := Wat2Wasm(`(module (table (export "") 1 funcref))`)
	assertNoError(err)
	tableModule, err := NewModule(store.Engine, tableWasm)
	assertNoError(err)
	instance, err := NewInstance(store, tableModule, []*Extern{})
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
	module, err := NewModule(store.Engine, wasm)
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

func ExampleLinker() {
	store := NewStore(NewEngine())

	// Compile two wasm modules where the first references the second
	wasm1, err := Wat2Wasm(`
	    (module
		(import "wasm2" "double" (func $double (param i32) (result i32)))
		(func (export "double_and_add") (param i32 i32) (result i32)
		  local.get 0
		  call $double
		  local.get 1
		  i32.add)
	    )
	`)
	check(err)

	wasm2, err := Wat2Wasm(`
	    (module
		(func (export "double") (param i32) (result i32)
		  local.get 0
		  i32.const 2
		  i32.mul)
	    )
	`)
	check(err)

	// Next compile both modules
	module1, err := NewModule(store.Engine, wasm1)
	check(err)
	module2, err := NewModule(store.Engine, wasm2)
	check(err)

	linker := NewLinker(store)

	// The second module is instantiated first since it has no imports, and
	// then we insert the instance back into the linker under the name
	// the first module expects.
	instance2, err := linker.Instantiate(module2)
	check(err)
	err = linker.DefineInstance("wasm2", instance2)
	check(err)

	// And now we can instantiate our first module, executing the result
	// afterwards
	instance1, err := linker.Instantiate(module1)
	check(err)
	doubleAndAdd := instance1.GetExport("double_and_add").Func()
	result, err := doubleAndAdd.Call(2, 3)
	check(err)
	fmt.Print(result.(int32))
	// Output: 7
}

func TestLinkerModule(t *testing.T) {
	store := NewStore(NewEngine())
	wasm, err := Wat2Wasm(`(module
	  (func (export "f"))
	)`)
	assertNoError(err)
	module, err := NewModule(store.Engine, wasm)
	assertNoError(err)

	linker := NewLinker(store)
	err = linker.DefineModule("foo", module)
	assertNoError(err)

	wasm, err = Wat2Wasm(`(module
	  (import "foo" "f" (func))
	)`)
	assertNoError(err)
	module, err = NewModule(store.Engine, wasm)
	assertNoError(err)

	_, err = linker.Instantiate(module)
	assertNoError(err)
}

func TestLinkerGetDefault(t *testing.T) {
	store := NewStore(NewEngine())
	linker := NewLinker(store)
	f, err := linker.GetDefault("foo")
	assertNoError(err)
	f.Call()
}

func TestLinkerGetOneByName(t *testing.T) {
	store := NewStore(NewEngine())
	linker := NewLinker(store)
	f, err := linker.GetOneByName("foo", "bar")
	if f != nil {
		panic("expected nil")
	}
	if err == nil {
		panic("expected an error")
	}

	err = linker.DefineFunc("foo", "baz", func() {})
	assertNoError(err)
	f, err = linker.GetOneByName("foo", "baz")
	assertNoError(err)
	f.Func().Call()
}
