//go:build wasmtime_component_model

package wasmtime

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

const hostFunctionsComponent = `(component
	(import "host-resource" (type $host-resource (sub resource)))
  (import "host-add" (func $host-add (param "value" u32)))
  (import "host-fail" (func $host-fail))
  (import "host-panic" (func $host-panic))
  (core func $host-add-lowered (canon lower (func $host-add)))
  (core func $host-fail-lowered (canon lower (func $host-fail)))
  (core func $host-panic-lowered (canon lower (func $host-panic)))
  (core module $m
    (import "" "host-add" (func $host-add (param i32)))
    (import "" "host-fail" (func $host-fail))
    (import "" "host-panic" (func $host-panic))
    (func (export "call-host") (param i32) (call $host-add (local.get 0)))
    (func (export "call-fail") (call $host-fail))
    (func (export "call-panic") (call $host-panic)))
  (core instance $i (instantiate $m
    (with "" (instance
      (export "host-add" (func $host-add-lowered))
      (export "host-fail" (func $host-fail-lowered))
      (export "host-panic" (func $host-panic-lowered))))))
  (func (export "call-host") (param "value" u32) (canon lift (core func $i "call-host")))
  (func (export "call-fail") (canon lift (core func $i "call-fail")))
  (func (export "call-panic") (canon lift (core func $i "call-panic"))))`

type componentHostState struct {
	calls int
	last  uint32
}

func TestComponentHostFunctionsAndLinkerBorrow(t *testing.T) {
	engine := newComponentEngine()
	defer engine.Close()
	component := newComponent(t, engine, hostFunctionsComponent)
	defer component.Close()
	state := &componentHostState{}
	store := NewStoreWithData(engine, state)
	linker := NewComponentLinker(engine)
	defer linker.Close()
	root := linker.Root()
	require.Panics(t, func() { linker.Instantiate(store, component) })
	require.Panics(t, func() { linker.Close() })
	require.Error(t, root.AddFunc("nil-host", nil))
	require.NoError(t, root.AddFunc("host-add", func(caller *ComponentCaller, args []*ComponentVal) ([]*ComponentVal, error) {
		hostState := caller.Data().(*componentHostState)
		hostState.calls++
		hostState.last = args[0].Value().(uint32)
		return nil, nil
	}))
	require.NoError(t, root.AddFunc("host-fail", func(*ComponentCaller, []*ComponentVal) ([]*ComponentVal, error) {
		return nil, errors.New("expected host failure")
	}))
	require.NoError(t, root.AddFunc("host-panic", func(*ComponentCaller, []*ComponentVal) ([]*ComponentVal, error) {
		panic("expected host panic")
	}))
	resourceType := NewComponentResourceType(7)
	defer resourceType.Close()
	require.Error(t, root.AddResource("nil-resource", resourceType, nil))
	require.NoError(t, root.AddResource("host-resource", resourceType, func(*ComponentCaller, uint32) error { return nil }))
	root.Close()

	instance, err := linker.Instantiate(store, component)
	require.NoError(t, err)
	callHost := instance.GetFunc(store, nil, "call-host")
	require.NotNil(t, callHost)
	argument := NewComponentU32(42)
	results, err := callHost.Call(store, []*ComponentVal{argument})
	argument.Close()
	require.NoError(t, err)
	require.Empty(t, results)
	require.Equal(t, &componentHostState{calls: 1, last: 42}, state)

	callPanic := instance.GetFunc(store, nil, "call-panic")
	require.PanicsWithValue(t, "expected host panic", func() {
		_, _ = callPanic.Call(store, nil)
	})

	failureStore := NewStoreWithData(engine, &componentHostState{})
	failureInstance, err := linker.Instantiate(failureStore, component)
	require.NoError(t, err)
	callFail := failureInstance.GetFunc(failureStore, nil, "call-fail")
	_, err = callFail.Call(failureStore, nil)
	require.ErrorContains(t, err, "expected host failure")
}

func TestComponentHostResourceOwnership(t *testing.T) {
	engine := newComponentEngine()
	defer engine.Close()
	store := NewStore(engine)
	first := NewComponentResourceType(11)
	clone := first.Clone()
	other := NewComponentResourceType(12)
	require.True(t, first.Equal(clone))
	require.False(t, first.Equal(other))
	first.Close()
	clone.Close()
	other.Close()

	value, err := NewComponentHostResource(store, ComponentHostResource{Owned: true, Rep: 42, TypeID: 11})
	require.NoError(t, err)
	require.Equal(t, ComponentValKindResource, value.Kind())
	resource, err := value.TakeHostResource(store)
	require.NoError(t, err)
	require.Equal(t, ComponentHostResource{Owned: true, Rep: 42, TypeID: 11}, resource)
	value.Close()

	dropped, err := NewComponentHostResource(store, ComponentHostResource{Owned: true, Rep: 43, TypeID: 11})
	require.NoError(t, err)
	require.NoError(t, dropped.DropResource(store))
	dropped.Close()
}
