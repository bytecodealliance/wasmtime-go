package wasmtime

import "testing"

func TestTable(t *testing.T) {
	store := NewStore(NewEngine())
	ty := NewTableType(NewValType(KindFuncref), Limits{Min: 1, Max: 3})
	table, err := NewTable(store, ty, nil)
	if err != nil {
		panic(err)
	}
	if table.Size() != 1 {
		panic("wrong size")
	}

	f, err := table.Get(0)
	if err != nil {
		panic(err)
	}
	if f != nil {
		panic("expected nil")
	}
	f, err = table.Get(1)
	if err == nil {
		panic("expected error")
	}
	if f != nil {
		panic("expected nil")
	}

	err = table.Set(0, nil)
	if err != nil {
		panic(err)
	}
	err = table.Set(1, nil)
	if err == nil {
		panic("expected error")
	}
	err = table.Set(0, WrapFunc(store, func() {}))
	if err != nil {
		panic(nil)
	}
	f, err = table.Get(0)
	if err != nil {
		panic(err)
	}
	if f == nil {
		panic("expected not nil")
	}

	prevSize, err := table.Grow(1, nil)
	if err != nil {
		panic(err)
	}
	if prevSize != 1 {
		print(prevSize)
		panic("bad prev")
	}
	f, err = table.Get(1)
	if err != nil {
		panic(err)
	}
	if f != nil {
		panic("expected nil")
	}

	called := false
	_, err = table.Grow(1, WrapFunc(store, func() {
		called = true
	}))
	if err != nil {
		panic(err)
	}
	f, err = table.Get(2)
	if err != nil {
		panic(err)
	}
	if called {
		panic("called already?")
	}
	_, err = f.Call()
	if err != nil {
		panic(err)
	}
	if !called {
		panic("should have called")
	}
}
