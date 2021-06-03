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
	instance, err := NewInstance(store, module, []AsExtern{})
	if err != nil {
		panic(err)
	}
	return instance, store
}

func TestRefTypesSmoke(t *testing.T) {
	instance, store := refTypesInstance(`
(module
  (func (export "f") (param externref) (result externref)
    local.get 0
  )
  (func (export "null_externref") (result externref)
    ref.null extern
  )
)
`)

	null_externref := instance.GetExport(store, "null_externref").Func()
	result, err := null_externref.Call(store)
	if err != nil {
		panic(err)
	}
	if result != nil {
		panic("expected nil result")
	}

	f := instance.GetExport(store, "f").Func()
	result, err = f.Call(store, true)
	if err != nil {
		panic(err)
	}
	if result.(bool) != true {
		panic("expected `true`")
	}

	result, err = f.Call(store, "x")
	if err != nil {
		panic(err)
	}
	if result.(string) != "x" {
		panic("expected `x`")
	}

	result, err = f.Call(store, ValExternref("x"))
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
		val, err := table.Get(store, uint32(i))
		if err != nil {
			panic(err)
		}
		if val.Get().(string) != "init" {
			panic("bad init")
		}
	}

	_, err = table.Grow(store, 2, ValExternref("grown"))
	if err != nil {
		panic(err)
	}
	for i := 0; i < 10; i++ {
		val, err := table.Get(store, uint32(i))
		if err != nil {
			panic(err)
		}
		if val.Get().(string) != "init" {
			panic("bad init")
		}
	}
	for i := 10; i < 12; i++ {
		val, err := table.Get(store, uint32(i))
		if err != nil {
			panic(err)
		}
		if val.Get().(string) != "grown" {
			panic("bad init")
		}
	}

	err = table.Set(store, 7, ValExternref("lucky"))
	if err != nil {
		panic(err)
	}

	for i := 0; i < 7; i++ {
		val, err := table.Get(store, uint32(i))
		if err != nil {
			panic(err)
		}
		if val.Get().(string) != "init" {
			panic("bad init")
		}
	}
	val, err := table.Get(store, 7)
	if err != nil {
		panic(err)
	}
	if val.Get().(string) != "lucky" {
		panic("bad init")
	}
	for i := 8; i < 10; i++ {
		val, err := table.Get(store, uint32(i))
		if err != nil {
			panic(err)
		}
		if val.Get().(string) != "init" {
			panic("bad init")
		}
	}
	for i := 10; i < 12; i++ {
		val, err := table.Get(store, uint32(i))
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

	val := global.Get(store)
	if val.Get().(string) != "hello" {
		panic("bad init")
	}
	err = global.Set(store, ValExternref("goodbye"))
	if err != nil {
		panic(err)
	}
	if global.Get(store).Get().(string) != "goodbye" {
		panic("bad init")
	}
}

func TestRefTypesWrap(t *testing.T) {
	store := refTypesStore()
	f := WrapFunc(store, func() error {
		return nil
	})
	if len(f.Type(store).Params()) != 0 {
		panic("wrong params")
	}
	if len(f.Type(store).Results()) != 1 {
		panic("wrong results")
	}
	if f.Type(store).Results()[0].Kind() != KindExternref {
		panic("wrong result")
	}
	ret, err := f.Call(store)
	if err != nil {
		panic(err)
	}
	if ret != nil {
		panic("expected nil error")
	}

	f = WrapFunc(store, func() error {
		return errors.New("message")
	})
	ret, err = f.Call(store)
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
	if len(f.Type(store).Params()) != 4 {
		panic("wrong params")
	}
	if f.Type(store).Params()[0].Kind() != KindExternref {
		panic("wrong param")
	}
	if f.Type(store).Params()[1].Kind() != KindExternref {
		panic("wrong param")
	}
	if f.Type(store).Params()[2].Kind() != KindFuncref {
		panic("wrong param")
	}
	if f.Type(store).Params()[3].Kind() != KindExternref {
		panic("wrong param")
	}
	if len(f.Type(store).Results()) != 2 {
		panic("wrong results")
	}
	if f.Type(store).Results()[0].Kind() != KindExternref {
		panic("wrong result")
	}
	if f.Type(store).Results()[1].Kind() != KindFuncref {
		panic("wrong result")
	}

	ret, err = f.Call(store, "message", store, f, errors.New("x"))
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
	gc := func() *GcHit {
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
		global.Set(store, ValExternref(obj))
		runtime.GC()
		if gc.hit {
			panic("gc too early")
		}
		global.Set(store, ValExternref(nil))
		return gc
	}()

	// Try real hard to get the Go GC to run the destructor. This is
	// somewhat nondeterministic depending on the platform, so this just
	// hits `runtime.GC()` a lot and hopes that eventually it actually
	// cleans up the `obj` from above. If this loop runs too many times,
	// though, we assume it'll never work and we fail the test.
	for i := 0; ; i++ {
		runtime.GC()
		if gc.hit {
			break
		}
		if i >= 10000 {
			panic("dtor not run")
		}
	}
}

func TestFuncFinalizer(t *testing.T) {
	instance, store := refTypesInstance(`
	      (module (func (export "f") (param externref)))
	`)
	f := instance.GetExport(store, "f").Func()
	obj, gc := newObjToDrop()
	_, err := f.Call(store, obj)
	if err != nil {
		panic(err)
	}
	store.GC()
	// like above, try real hard to get the Go GC to run the destructor
	for i := 0; ; i++ {
		runtime.GC()
		if gc.hit {
			break
		}
		if i >= 10000 {
			panic("dtor not run")
		}
	}
}
