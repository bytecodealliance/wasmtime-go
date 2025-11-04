package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	engine := NewEngine()
	defer engine.Close()
	store := NewStore(engine)
	defer store.Close()
}

func TestInterruptWasm(t *testing.T) {
	config := NewConfig()
	config.SetEpochInterruption(true)
	store := NewStore(NewEngineWithConfig(config))
	store.SetEpochDeadline(1)
	wasm, err := Wat2Wasm(`
	  (import "" "" (func))
	  (func
	    call 0
	    (loop br 0))
	  (start 1)
	`)
	require.NoError(t, err)
	module, err := NewModule(store.Engine, wasm)
	require.NoError(t, err)
	engine := store.Engine
	f := WrapFunc(store, func() {
		engine.IncrementEpoch()
	})
	instance, err := NewInstance(store, module, []AsExtern{f})
	require.Nil(t, instance)

	require.Error(t, err)
	trap := err.(*Trap)
	require.NotNil(t, trap)
}

func TestFuelConsumed(t *testing.T) {
	engine := NewEngine()
	store := NewStore(engine)

	fuel, enable := store.GetFuel()
	require.Error(t, enable)
	require.Equal(t, fuel, uint64(0))
}

func TestAddFuel(t *testing.T) {
	config := NewConfig()
	config.SetConsumeFuel(true)
	engine := NewEngineWithConfig(config)
	store := NewStore(engine)

	fuel, enable := store.GetFuel()
	require.NoError(t, enable)
	require.Equal(t, fuel, uint64(0))

	const add_fuel = 3
	err := store.SetFuel(add_fuel)
	require.NoError(t, err)
}

func TestLimiterMemorySizeFail(t *testing.T) {
	engine := NewEngine()
	store := NewStore(engine)

	store.Limiter(2*64*1024, -1, -1, -1, -1)
	wasm, err := Wat2Wasm(`
	(module
	  (memory 3)
	)
	`)
	require.NoError(t, err)

	module, err := NewModule(store.Engine, wasm)
	require.NoError(t, err)

	_, err = NewInstance(store, module, []AsExtern{})
	require.Error(t, err, "memory minimum size of 3 pages exceeds memory limits")
}

func TestLimiterMemorySizeSuccess(t *testing.T) {
	engine := NewEngine()
	store := NewStore(engine)

	store.Limiter(4*64*1024, -1, -1, -1, -1)
	wasm, err := Wat2Wasm(`
	(module
	  (memory 3)
	)
	`)
	require.NoError(t, err)

	module, err := NewModule(store.Engine, wasm)
	require.NoError(t, err)

	_, err = NewInstance(store, module, []AsExtern{})
	require.NoError(t, err)
}

type embeddedStoreData []string

type testStoreData struct {
	s string
	i int32

	m   map[string]any
	ptr *embeddedStoreData

	inline struct {
		foo string
	}
}

func TestStoreData(t *testing.T) {
	embedded := embeddedStoreData{"what", "ever"}
	data := &testStoreData{
		s: "foobar",
		i: 42,
		m: map[string]any{
			"intval":    42,
			"stringval": "hello",
		},
		ptr:    &embedded,
		inline: struct{ foo string }{"bar"},
	}

	engine := NewEngine()
	store := NewStoreWithData(engine, data)
	wasm, err := Wat2Wasm(`
	  (module
	  	(func $f1 (import "" "f1"))
		(func (export "run") (call $f1))
      )
	`)
	require.NoError(t, err)

	module, err := NewModule(store.Engine, wasm)
	require.NoError(t, err)

	called := false
	f := WrapFunc(store, func(c *Caller) {
		t.Helper()
		called = true
		sd, ok := c.Data().(*testStoreData)
		require.True(t, ok)
		require.Equal(t, data, sd)
	})

	instance, err := NewInstance(store, module, []AsExtern{f})

	_, err = instance.GetFunc(store, "run").Call(store)
	require.NoError(t, err)
	require.True(t, called, "expected wrapped func to be called")
}
