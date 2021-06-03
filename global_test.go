package wasmtime

import "testing"

func TestGlobal(t *testing.T) {
	store := NewStore(NewEngine())
	g, err := NewGlobal(store, NewGlobalType(NewValType(KindI32), true), ValI32(100))
	if err != nil {
		panic(err)
	}
	if g.Get(store).I32() != 100 {
		panic("wrong value in global")
	}
	g.Set(store, ValI32(200))
	if g.Get(store).I32() != 200 {
		panic("wrong value in global")
	}

	_, err = NewGlobal(store, NewGlobalType(NewValType(KindI64), true), ValI32(100))
	if err == nil {
		panic("should fail to create global")
	}
	err = g.Set(store, ValI64(200))
	if err == nil {
		panic("should fail to set global")
	}
}
