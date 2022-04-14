package wasmtime

import (
	"errors"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func refTypesStore() *Store {
	config := NewConfig()
	config.SetWasmReferenceTypes(true)
	return NewStore(NewEngineWithConfig(config))
}

func refTypesInstance(t *testing.T, wat string) (*Instance, *Store) {
	store := refTypesStore()
	wasm, err := Wat2Wasm(wat)
	assert.NoError(t, err)
	module, err := NewModule(store.Engine, wasm)
	assert.NoError(t, err)
	instance, err := NewInstance(store, module, []AsExtern{})
	assert.NoError(t, err)
	return instance, store
}

func TestRefTypesSmoke(t *testing.T) {
	instance, store := refTypesInstance(t, `
(module
  (func (export "f") (param externref) (result externref)
    local.get 0
  )
  (func (export "null_externref") (result externref)
    ref.null extern
  )
)
`)

	null_externref := instance.GetFunc(store, "null_externref")
	result, err := null_externref.Call(store)
	assert.NoError(t, err)
	if result != nil {
		panic("expected nil result")
	}

	f := instance.GetFunc(store, "f")
	result, err = f.Call(store, true)
	assert.NoError(t, err)
	assert.True(t, result.(bool))

	result, err = f.Call(store, "x")
	assert.NoError(t, err)
	assert.Equal(t, "x", result.(string))

	result, err = f.Call(store, ValExternref("x"))
	assert.NoError(t, err)
	assert.Equal(t, "x", result.(string))
}

func TestRefTypesVal(t *testing.T) {
	val := ValExternref("x")
	assert.Equal(t, "x", val.Get().(string))
}

func TestRefTypesTable(t *testing.T) {
	store := refTypesStore()
	table, err := NewTable(
		store,
		NewTableType(NewValType(KindExternref), 10, false, 0),
		ValExternref("init"),
	)
	assert.NoError(t, err)

	for i := 0; i < 10; i++ {
		val, err := table.Get(store, uint32(i))
		assert.NoError(t, err)
		assert.Equal(t, "init", val.Get().(string))
	}

	_, err = table.Grow(store, 2, ValExternref("grown"))
	assert.NoError(t, err)
	for i := 0; i < 10; i++ {
		val, err := table.Get(store, uint32(i))
		assert.NoError(t, err)
		assert.Equal(t, "init", val.Get().(string))
	}
	for i := 10; i < 12; i++ {
		val, err := table.Get(store, uint32(i))
		assert.NoError(t, err)
		assert.Equal(t, "grown", val.Get().(string))
	}

	err = table.Set(store, 7, ValExternref("lucky"))
	assert.NoError(t, err)

	for i := 0; i < 7; i++ {
		val, err := table.Get(store, uint32(i))
		assert.NoError(t, err)
		assert.Equal(t, "init", val.Get().(string))
	}
	val, err := table.Get(store, 7)
	assert.NoError(t, err)
	assert.Equal(t, "lucky", val.Get().(string))

	for i := 8; i < 10; i++ {
		val, err := table.Get(store, uint32(i))
		assert.NoError(t, err)
		assert.Equal(t, "init", val.Get().(string))
	}
	for i := 10; i < 12; i++ {
		val, err := table.Get(store, uint32(i))
		assert.NoError(t, err)
		assert.Equal(t, "grown", val.Get().(string))
	}
}

func TestRefTypesGlobal(t *testing.T) {
	store := refTypesStore()
	global, err := NewGlobal(
		store,
		NewGlobalType(NewValType(KindExternref), true),
		ValExternref("hello"),
	)
	assert.NoError(t, err)

	val := global.Get(store)
	assert.Equal(t, "hello", val.Get().(string))
	err = global.Set(store, ValExternref("goodbye"))
	assert.NoError(t, err)
	assert.Equal(t, "goodbye", global.Get(store).Get().(string))
}

func TestRefTypesWrap(t *testing.T) {
	store := refTypesStore()
	f := WrapFunc(store, func() error {
		return nil
	})
	assert.Len(t, f.Type(store).Params(), 0)
	assert.Len(t, f.Type(store).Results(), 1)
	assert.Equal(t, KindExternref, f.Type(store).Results()[0].Kind())

	ret, err := f.Call(store)
	assert.NoError(t, err)
	assert.Nil(t, ret)

	expectedErr := errors.New("message")
	f = WrapFunc(store, func() error {
		return expectedErr
	})
	ret, err = f.Call(store)
	assert.NoError(t, err)
	assert.ErrorIs(t, ret.(error), expectedErr)

	f = WrapFunc(store, func(a interface{}, b *Store, f *Func, g error) (error, *Func) {
		assert.Equal(t, "message", a.(string))
		assert.Equal(t, "x", g.Error())

		return g, f
	})
	assert.Len(t, f.Type(store).Params(), 4)
	assert.Equal(t, KindExternref, f.Type(store).Params()[0].Kind())
	assert.Equal(t, KindExternref, f.Type(store).Params()[1].Kind())
	assert.Equal(t, KindFuncref, f.Type(store).Params()[2].Kind())
	assert.Equal(t, KindExternref, f.Type(store).Params()[3].Kind())

	assert.Len(t, f.Type(store).Results(), 2)
	assert.Equal(t, KindExternref, f.Type(store).Results()[0].Kind())
	assert.Equal(t, KindFuncref, f.Type(store).Results()[1].Kind())

	ret, err = f.Call(store, "message", store, f, errors.New("x"))
	assert.NoError(t, err)
	arr := ret.([]Val)
	assert.Len(t, arr, 2)
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
		assert.NoError(t, err)
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
		assert.Less(t, i, 10000)
	}
}

func TestFuncFinalizer(t *testing.T) {
	instance, store := refTypesInstance(t, `
	      (module (func (export "f") (param externref)))
	`)
	f := instance.GetFunc(store, "f")
	obj, gc := newObjToDrop()
	_, err := f.Call(store, obj)
	assert.NoError(t, err)
	store.GC()
	// like above, try real hard to get the Go GC to run the destructor
	for i := 0; ; i++ {
		runtime.GC()
		if gc.hit {
			break
		}
		assert.Less(t, i, 10000)
	}
}
