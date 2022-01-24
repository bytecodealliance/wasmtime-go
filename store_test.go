package wasmtime

import "testing"

func TestStore(t *testing.T) {
	engine := NewEngine()
	NewStore(engine)
}

func TestInterruptHandle(t *testing.T) {
	store := NewStore(NewEngine())
	handle, err := store.InterruptHandle()
	if handle != nil {
		panic("expected nil handle")
	}
	if err == nil {
		panic("expected an error")
	}

	config := NewConfig()
	config.SetInterruptable(true)
	store = NewStore(NewEngineWithConfig(config))
	handle, err = store.InterruptHandle()
	if err != nil {
		panic(err)
	}
	handle.Interrupt()
}

func TestInterruptWasm(t *testing.T) {
	config := NewConfig()
	config.SetInterruptable(true)
	store := NewStore(NewEngineWithConfig(config))
	handle, err := store.InterruptHandle()
	if err != nil {
		panic(err)
	}
	wasm, err := Wat2Wasm(`
	  (import "" "" (func))
	  (func
	    call 0
	    (loop br 0))
	  (start 1)
	`)
	if err != nil {
		panic(err)
	}
	module, err := NewModule(store.Engine, wasm)
	if err != nil {
		panic(err)
	}
	f := WrapFunc(store, func() {
		handle.Interrupt()
	})
	instance, err := NewInstance(store, module, []AsExtern{f})
	if instance != nil {
		panic("expected nil instance")
	}
	if err == nil {
		panic("expected an error")
	}
	trap := err.(*Trap)
	if trap == nil {
		panic("expected a trap")
	}
}

func TestFuelConsumed(t *testing.T) {
	engine := NewEngine()
	store := NewStore(engine)

	fuel, enable := store.FuelConsumed()
	if enable != false {
		t.Fatal("expected not enable")
	}
	if fuel != 0 {
		t.Fatalf("fuel is %d, not zero", fuel)
	}
}

func TestAddFuel(t *testing.T) {
	config := NewConfig()
	config.SetConsumeFuel(true)
	engine := NewEngineWithConfig(config)
	store := NewStore(engine)

	fuel, enable := store.FuelConsumed()
	if enable != true {
		t.Fatal("expected enabled")
	}
	if fuel != 0 {
		t.Fatalf("fuel is %d, not zero", fuel)
	}

	const add_fuel = 3
	if err := store.AddFuel(add_fuel); err != nil {
		t.Fatal("expected no error")
	}
}

func TestConsumeFuel(t *testing.T) {
	config := NewConfig()
	config.SetConsumeFuel(true)
	engine := NewEngineWithConfig(config)
	store := NewStore(engine)

	fuel, enable := store.FuelConsumed()
	if enable != true {
		t.Fatal("expected enabled")
	}
	if fuel != 0 {
		t.Fatalf("fuel is %d, not zero", fuel)
	}

	const add_fuel = 3
	if err := store.AddFuel(add_fuel); err != nil {
		t.Fatal("expected no error")
	}

	consume_fuel := uint64(1)
	remaining, err := store.ConsumeFuel(consume_fuel)
	if err != nil {
		t.Fatal("expected no error")
	}
	if remaining != (add_fuel - consume_fuel) {
		t.Fatalf("expected %d, but %d", add_fuel-consume_fuel, remaining)
	}
}
