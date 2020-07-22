package wasmtime

import (
	"errors"
	"runtime"
	"testing"
)

func refTypesStore() *Store {
	config := NewConfig()
	config.SetWasmReferenceTypes(true)
	return NewStore(NewEngineWithConfig(config))
}

func refTypesInstance(wat string) (*Instance, *Store) {
	store := refTypesStore()
	wasm, err := Wat2Wasm(wat)
	if err != nil {
		panic(err)
	}
	module, err := NewModule(store.Engine, wasm)
	if err != nil {
		panic(err)
	}
	instance, err := NewInstance(store, module, []*Extern{})
	if err != nil {
		panic(err)
	}
	return instance, store
}

func TestRefTypesSmoke(t *testing.T) {
	instance, _ := refTypesInstance(`
(module
  (func (export "f") (param externref) (result externref)
    local.get 0
  )
  (func (export "null_externref") (result externref)
    ref.null extern
  )
)
`)

	null_externref := instance.GetExport("null_externref").Func()
	result, err := null_externref.Call()
	if err != nil {
		panic(err)
	}
	if result != nil {
		panic("expected nil result")
	}

	f := instance.GetExport("f").Func()
	result, err = f.Call(true)
	if err != nil {
		panic(err)
	}
	if result.(bool) != true {
		panic("expected `true`")
	}

	result, err = f.Call("x")
	if err != nil {
		panic(err)
	}
	if result.(string) != "x" {
		panic("expected `x`")
	}

	result, err = f.Call(ValExternref("x"))
	if err != nil {
		panic(err)
	}
	if result.(string) != "x" {
		panic("expected `x`")
	}
}

func TestRefTypesVal(t *testing.T) {
	val := ValExternref("x")
	if val.Get().(string) != "x" {
		panic("bad return of `Get`")
	}
}

func TestRefTypesTable(t *testing.T) {
	store := refTypesStore()
	table, err := NewTable(
		store,
		NewTableType(NewValType(KindExternref), Limits{Min: 10, Max: LimitsMaxNone}),
		ValExternref("init"),
	)
	if err != nil {
		panic("err")
	}

	for i := 0; i < 10; i++ {
		val, err := table.Get(uint32(i))
		if err != nil {
			panic(err)
		}
		if val.Get().(string) != "init" {
			panic("bad init")
		}
	}

	_, err = table.Grow(2, ValExternref("grown"))
	if err != nil {
		panic(err)
	}
	for i := 0; i < 10; i++ {
		val, err := table.Get(uint32(i))
		if err != nil {
			panic(err)
		}
		if val.Get().(string) != "init" {
			panic("bad init")
		}
	}
	for i := 10; i < 12; i++ {
		val, err := table.Get(uint32(i))
		if err != nil {
			panic(err)
		}
		if val.Get().(string) != "grown" {
			panic("bad init")
		}
	}

	err = table.Set(7, ValExternref("lucky"))
	if err != nil {
		panic(err)
	}

	for i := 0; i < 7; i++ {
		val, err := table.Get(uint32(i))
		if err != nil {
			panic(err)
		}
		if val.Get().(string) != "init" {
			panic("bad init")
		}
	}
	val, err := table.Get(7)
	if err != nil {
		panic(err)
	}
	if val.Get().(string) != "lucky" {
		panic("bad init")
	}
	for i := 8; i < 10; i++ {
		val, err := table.Get(uint32(i))
		if err != nil {
			panic(err)
		}
		if val.Get().(string) != "init" {
			panic("bad init")
		}
	}
	for i := 10; i < 12; i++ {
		val, err := table.Get(uint32(i))
		if err != nil {
			panic(err)
		}
		if val.Get().(string) != "grown" {
			panic("bad init")
		}
	}
}

func TestRefTypesGlobal(t *testing.T) {
	store := refTypesStore()
	global, err := NewGlobal(
		store,
		NewGlobalType(NewValType(KindExternref), true),
		ValExternref("hello"),
	)
	if err != nil {
		panic(err)
	}

	val := global.Get()
	if val.Get().(string) != "hello" {
		panic("bad init")
	}
	err = global.Set(ValExternref("goodbye"))
	if err != nil {
		panic(err)
	}
	if global.Get().Get().(string) != "goodbye" {
		panic("bad init")
	}
}

func TestRefTypesWrap(t *testing.T) {
	store := refTypesStore()
	f := WrapFunc(store, func() error {
		return nil
	})
	if len(f.Type().Params()) != 0 {
		panic("wrong params")
	}
	if len(f.Type().Results()) != 1 {
		panic("wrong results")
	}
	if f.Type().Results()[0].Kind() != KindExternref {
		panic("wrong result")
	}
	ret, err := f.Call()
	if err != nil {
		panic(err)
	}
	if ret != nil {
		panic("expected nil error")
	}

	f = WrapFunc(store, func() error {
		return errors.New("message")
	})
	ret, err = f.Call()
	if err != nil {
		panic(err)
	}
	if ret.(error).Error() != "message" {
		panic(ret.(error))
	}

	f = WrapFunc(store, func(a interface{}, b *Store, f *Func, g error) (error, *Func) {
		if a.(string) != "message" {
			panic("bad message")
		}
		if g.Error() != "x" {
			panic("bad error")
		}
		return g, f
	})
	if len(f.Type().Params()) != 4 {
		panic("wrong params")
	}
	if f.Type().Params()[0].Kind() != KindExternref {
		panic("wrong param")
	}
	if f.Type().Params()[1].Kind() != KindExternref {
		panic("wrong param")
	}
	if f.Type().Params()[2].Kind() != KindFuncref {
		panic("wrong param")
	}
	if f.Type().Params()[3].Kind() != KindExternref {
		panic("wrong param")
	}
	if len(f.Type().Results()) != 2 {
		panic("wrong results")
	}
	if f.Type().Results()[0].Kind() != KindExternref {
		panic("wrong result")
	}
	if f.Type().Results()[1].Kind() != KindFuncref {
		panic("wrong result")
	}

	ret, err = f.Call("message", store, f, errors.New("x"))
	if err != nil {
		panic(err)
	}
	arr := ret.([]Val)
	if len(arr) != 2 {
		panic("bad ret")
	}
}

type GcHit struct {
	hit bool
}

type ObjToDrop struct {
	ptr *GcHit
}

func newObjToDrop() (*ObjToDrop, *GcHit) {
	gc := &GcHit{hit: false}
	obj := &ObjToDrop{ptr: gc}
	runtime.SetFinalizer(obj, func(obj *ObjToDrop) {
		obj.ptr.hit = true
	})
	return obj, gc
}

func TestGlobalFinalizer(t *testing.T) {
	store := refTypesStore()
	global, err := NewGlobal(
		store,
		NewGlobalType(NewValType(KindExternref), true),
		ValExternref(nil),
	)
	if err != nil {
		panic(err)
	}
	obj, gc := newObjToDrop()
	global.Set(ValExternref(obj))
	runtime.GC()
	if gc.hit {
		panic("gc too early")
	}
	global.Set(ValExternref(nil))
	runtime.GC()
	if !gc.hit {
		panic("dtor not run")
	}
}

func TestFuncFinalizer(t *testing.T) {
	instance, store := refTypesInstance(`
	      (module (func (export "f") (param externref)))
	`)
	f := instance.GetExport("f").Func()
	obj, gc := newObjToDrop()
	_, err := f.Call(obj)
	if err != nil {
		panic(err)
	}
	store.GC()
	runtime.GC()
	if !gc.hit {
		panic("dtor not run")
	}
}
