package wasmtime

import "testing"

func moduleLinkingStore() *Store {
	config := NewConfig()
	config.SetWasmModuleLinking(true)
	return NewStore(NewEngineWithConfig(config))
}

func TestModuleType(t *testing.T) {
	wasm, err := Wat2Wasm(`
          (module
            (import "" "f" (func))
            (import "a" "g" (global i32))
            (import "b" (table 1 funcref))
            (import "c" "" (memory 1))

	    (func (export "y"))
	    (global (export "z") i32 (i32.const 0))
	    (table (export "x") 1 funcref)
          )
        `)
	if err != nil {
		panic(err)
	}
	store := moduleLinkingStore()
	module, err := NewModule(store.Engine, wasm)
	if err != nil {
		panic(err)
	}
	ty := module.Type()
	imports := ty.Imports()
	if len(imports) != 4 {
		panic("wrong number of imports")
	}
	if imports[2].Name() != nil {
		panic("bad import name")
	}
	exports := ty.Exports()
	if len(exports) != 3 {
		panic("wrong number of exports")
	}
}

func TestInstanceType(t *testing.T) {
	wasm, err := Wat2Wasm(`
          (module
	    (func (export "y"))
	    (global (export "z") i32 (i32.const 0))
	    (table (export "x") 1 funcref)
          )
        `)
	if err != nil {
		panic(err)
	}
	store := moduleLinkingStore()
	module, err := NewModule(store.Engine, wasm)
	if err != nil {
		panic(err)
	}
	instance, err := NewInstance(store, module, []*Extern{})
	if err != nil {
		panic(err)
	}
	ty := instance.Type()
	exports := ty.Exports()
	if len(exports) != 3 {
		panic("wrong number of exports")
	}
}

func TestImportModule(t *testing.T) {
	wasm, err := Wat2Wasm(`(module (import "" (module)))`)
	if err != nil {
		panic(err)
	}
	store := moduleLinkingStore()
	module, err := NewModule(store.Engine, wasm)
	if err != nil {
		panic(err)
	}
	wasm, err = Wat2Wasm(`(module)`)
	if err != nil {
		panic(err)
	}
	module2, err := NewModule(store.Engine, wasm)
	if err != nil {
		panic(err)
	}
	_, err = NewInstance(store, module, []*Extern{module2.AsExtern()})
	if err != nil {
		panic(err)
	}
}

func TestImportInstance(t *testing.T) {
	wasm, err := Wat2Wasm(`(module (import "" (instance)))`)
	if err != nil {
		panic(err)
	}
	store := moduleLinkingStore()
	module, err := NewModule(store.Engine, wasm)
	if err != nil {
		panic(err)
	}
	wasm, err = Wat2Wasm(`(module)`)
	if err != nil {
		panic(err)
	}
	module2, err := NewModule(store.Engine, wasm)
	if err != nil {
		panic(err)
	}
	instance, err := NewInstance(store, module2, []*Extern{})
	if err != nil {
		panic(err)
	}
	_, err = NewInstance(store, module, []*Extern{instance.AsExtern()})
	if err != nil {
		panic(err)
	}
}

func TestExportModule(t *testing.T) {
	wasm, err := Wat2Wasm(`(module (module (export "")))`)
	if err != nil {
		panic(err)
	}
	store := moduleLinkingStore()
	module, err := NewModule(store.Engine, wasm)
	if err != nil {
		panic(err)
	}
	instance, err := NewInstance(store, module, []*Extern{})
	if err != nil {
		panic(err)
	}
	if instance.GetExport("").Module() == nil {
		panic("expected a module")
	}
}

func TestExportInstance(t *testing.T) {
	wasm, err := Wat2Wasm(`(module (module) (instance (export "") (instantiate 0)))`)
	if err != nil {
		panic(err)
	}
	store := moduleLinkingStore()
	module, err := NewModule(store.Engine, wasm)
	if err != nil {
		panic(err)
	}
	instance, err := NewInstance(store, module, []*Extern{})
	if err != nil {
		panic(err)
	}
	if instance.GetExport("").Instance() == nil {
		panic("expected a module")
	}
}
