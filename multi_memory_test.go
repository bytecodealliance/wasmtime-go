package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
	require.NoError(t, err)
	store := multiMemoryStore()
	module, err := NewModule(store.Engine, wasm)
	require.NoError(t, err)
	exports := module.Exports()
	require.Len(t, exports, 2)
	require.NotNil(t, exports[0].Type().MemoryType())
	require.Equal(t, uint64(2), exports[0].Type().MemoryType().Minimum())

	present, max := exports[0].Type().MemoryType().Maximum()
	require.True(t, present, "wrong memory limits")
	require.Equal(t, uint64(3), max, "wrong memory limits")

	require.NotNil(t, exports[1].Type().MemoryType())
	require.Equal(t, uint64(2), exports[0].Type().MemoryType().Minimum())

	present, max = exports[1].Type().MemoryType().Maximum()
	require.True(t, present, "wrong memory limits")
	require.Equal(t, uint64(4), max, "wrong memory limits")

	_, err = NewInstance(store, module, nil)
	require.NoError(t, err)
}

func TestMultiMemoryImported(t *testing.T) {
	wasm, err := Wat2Wasm(`
    (module
      (import "" "m0" (memory 1))
      (import "" "m1" (memory $m 1))
      (func (export "load1") (result i32)
        i32.const 2
        i32.load8_s $m
      )
    )`)
	require.NoError(t, err)
	store := multiMemoryStore()

	mem0, err := NewMemory(store, NewMemoryType(1, true, 3, false))
	require.NoError(t, err)
	mem1, err := NewMemory(store, NewMemoryType(2, true, 4, false))
	require.NoError(t, err)

	module, err := NewModule(store.Engine, wasm)
	require.NoError(t, err)
	instance, err := NewInstance(store, module, []AsExtern{mem0, mem1})
	require.NoError(t, err)

	copy(mem1.UnsafeData(store)[2:3], []byte{100})

	res, err := instance.GetFunc(store, "load1").Call(store)
	require.NoError(t, err)
	require.IsType(t, res, int32(0))
	require.Equal(t, int32(100), res.(int32))
}
