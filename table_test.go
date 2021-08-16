package wasmtime

import "testing"

func TestTable(t *testing.T) {
	store := NewStore(NewEngine())
	ty := NewTableType(NewValType(KindFuncref), 1, true, 3)
	table, err := NewTable(store, ty, ValFuncref(nil))
	if err != nil {
		panic(err)
	}
	if table.Size(store) != 1 {
		panic("wrong size")
	}

	f, err := table.Get(store, 0)
	if err != nil {
		panic(err)
	}
	if f.Funcref() != nil {
		panic("expected nil")
	}
	f, err = table.Get(store, 1)
	if err == nil {
		panic("expected error")
	}

	err = table.Set(store, 0, ValFuncref(nil))
	if err != nil {
		panic(err)
	}
	err = table.Set(store, 1, ValFuncref(nil))
	if err == nil {
		panic("expected error")
	}
	err = table.Set(store, 0, ValFuncref(WrapFunc(store, func() {})))
	if err != nil {
		panic(nil)
	}
	f, err = table.Get(store, 0)
	if err != nil {
		panic(err)
	}
	if f.Funcref() == nil {
		panic("expected not nil")
	}

	prevSize, err := table.Grow(store, 1, ValFuncref(nil))
	if err != nil {
		panic(err)
	}
	if prevSize != 1 {
		print(prevSize)
		panic("bad prev")
	}
	f, err = table.Get(store, 1)
	if err != nil {
		panic(err)
	}
	if f.Funcref() != nil {
		panic("expected nil")
	}

	called := false
	_, err = table.Grow(store, 1, ValFuncref(WrapFunc(store, func() {
		called = true
	})))
	if err != nil {
		panic(err)
	}
	f, err = table.Get(store, 2)
	if err != nil {
		panic(err)
	}
	if called {
		panic("called already?")
	}
	_, err = f.Funcref().Call(store)
	if err != nil {
		panic(err)
	}
	if !called {
		panic("should have called")
	}
}
