package wasmtime

import "testing"

func multiMemoryStore() *Store {
	config := NewConfig()
	config.SetWasmMultiMemory(true)
	return NewStore(NewEngineWithConfig(config))
}

func TestMultiMemoryExported(t *testing.T) {
	wasm, err := Wat2Wasm(`
    (module
        (memory (export "memory0") 2 3)
        (memory (export "memory1") 2 4)
        (data (memory 0) (i32.const 0x1000) "\01\02\03\04")
        (data (memory 1) (i32.const 0x1000) "\04\03\02\01")
    )`)
	if err != nil {
		panic(err)
	}
	store := multiMemoryStore()
	module, err := NewModule(store.Engine, wasm)
	if err != nil {
		panic(err)
	}
	ty := module.Type()
	exports := ty.Exports()
	if len(exports) != 2 {
		panic("wrong number of exports")
	}
	if exports[0].Type().MemoryType() == nil {
		panic("wrong export type")
	}
	if (exports[0].Type().MemoryType().Limits() != Limits{Min: 2, Max: 3}) {
		panic("wrong memory limits")
	}
	if exports[1].Type().MemoryType() == nil {
		panic("wrong export type")
	}
	if (exports[1].Type().MemoryType().Limits() != Limits{Min: 2, Max: 4}) {
		panic("wrong memory limits")
	}

	_, err = NewInstance(store, module, nil)
	if err != nil {
		panic(err)
	}
}

func TestMultiMemoryImported(t *testing.T) {
	wasm, err := Wat2Wasm(`
    (module
      (import "" "m0" (memory 1))
      (import "" "m1" (memory 1))
      (func (export "load1") (result i32)
        i32.const 2
        i32.load8_s (memory 1)
      )
    )`)
	if err != nil {
		panic(err)
	}
	store := multiMemoryStore()

	mem0, err := NewMemory(store, NewMemoryType(Limits{Min: 1, Max: 3}))
	if err != nil {
		panic(err)
	}
	mem1, err := NewMemory(store, NewMemoryType(Limits{Min: 2, Max: 4}))
	if err != nil {
		panic(err)
	}

	module, err := NewModule(store.Engine, wasm)
	if err != nil {
		panic(err)
	}
	instance, err := NewInstance(store, module, []AsExtern{mem0, mem1})
	if err != nil {
		panic(err)
	}

	copy(mem1.UnsafeData(store)[2:3], []byte{100})

	res, err := instance.GetFunc(store, "load1").Call(store)
	if err != nil {
		panic(err)
	}
	if v, ok := res.(int32); !ok || v != 100 {
		panic("unexpected result")
	}
}
