package v2

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	engine := NewEngine()
	NewStore(engine)
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

	fuel, enable := store.FuelConsumed()
	require.False(t, enable)
	require.Equal(t, fuel, uint64(0))
}

func TestAddFuel(t *testing.T) {
	config := NewConfig()
	config.SetConsumeFuel(true)
	engine := NewEngineWithConfig(config)
	store := NewStore(engine)

	fuel, enable := store.FuelConsumed()
	require.True(t, enable)
	require.Equal(t, fuel, uint64(0))

	const add_fuel = 3
	err := store.AddFuel(add_fuel)
	require.NoError(t, err)
}

func TestConsumeFuel(t *testing.T) {
	config := NewConfig()
	config.SetConsumeFuel(true)
	engine := NewEngineWithConfig(config)
	store := NewStore(engine)

	fuel, enable := store.FuelConsumed()
	require.True(t, enable)
	require.Equal(t, fuel, uint64(0))

	const add_fuel = 3
	err := store.AddFuel(add_fuel)
	require.NoError(t, err)

	consume_fuel := uint64(1)
	remaining, err := store.ConsumeFuel(consume_fuel)
	require.NoError(t, err)
	require.Equal(t, (add_fuel - consume_fuel), remaining)
}
