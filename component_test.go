package wasmtime

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// newComponentEngine returns an Engine with the component-model proposal
// enabled, which is required to compile and instantiate components.
func newComponentEngine() *Engine {
	cfg := NewConfig()
	cfg.SetWasmComponentModel(true)
	return NewEngineWithConfig(cfg)
}

func TestComponent(t *testing.T) {
	engine := newComponentEngine()
	_, err := NewComponent(engine, []byte{})
	require.Error(t, err, "empty bytes should fail")
	_, err = NewComponent(engine, []byte{1, 2, 3})
	require.Error(t, err, "garbage bytes should fail")

	wasm, err := Wat2Wasm(`(component)`)
	require.NoError(t, err)
	component, err := NewComponent(engine, wasm)
	require.NoError(t, err)
	defer component.Close()
}

func TestComponentSerialize(t *testing.T) {
	engine := newComponentEngine()
	wasm, err := Wat2Wasm(`(component)`)
	require.NoError(t, err)
	component, err := NewComponent(engine, wasm)
	require.NoError(t, err)
	defer component.Close()

	bytes, err := component.Serialize()
	require.NoError(t, err)
	require.NotEmpty(t, bytes)

	round, err := NewComponentDeserialize(engine, bytes)
	require.NoError(t, err)
	defer round.Close()

	tmp, err := os.CreateTemp("", "component-serialize")
	require.NoError(t, err)
	defer os.Remove(tmp.Name())
	_, err = tmp.Write(bytes)
	require.NoError(t, err)
	require.NoError(t, tmp.Close())

	roundFromFile, err := NewComponentDeserializeFile(engine, tmp.Name())
	require.NoError(t, err)
	defer roundFromFile.Close()
}

func TestComponentInstantiate(t *testing.T) {
	engine := newComponentEngine()
	store := NewStore(engine)

	wasm, err := Wat2Wasm(`
      (component
        (core module $m
          (func (export "hello"))
        )
        (core instance $i (instantiate $m))
        (func (export "hello") (canon lift (core func $i "hello")))
      )
    `)
	require.NoError(t, err)
	component, err := NewComponent(engine, wasm)
	require.NoError(t, err)
	defer component.Close()

	linker := NewComponentLinker(engine)
	defer linker.Close()

	instance, err := linker.Instantiate(store, component)
	require.NoError(t, err)
	require.NotNil(t, instance)
}

func TestComponentDefineUnknownImportsAsTraps(t *testing.T) {
	engine := newComponentEngine()
	store := NewStore(engine)

	wasm, err := Wat2Wasm(`
      (component
        (import "host:missing/api" (instance
          (export "f" (func))
        ))
      )
    `)
	require.NoError(t, err)
	component, err := NewComponent(engine, wasm)
	require.NoError(t, err)
	defer component.Close()

	linker := NewComponentLinker(engine)
	defer linker.Close()

	// Without trap definitions the missing import causes instantiation to fail.
	_, err = linker.Instantiate(store, component)
	require.Error(t, err, "expected missing import to fail instantiation")

	// After defining unknown imports as traps the instantiation succeeds.
	require.NoError(t, linker.DefineUnknownImportsAsTraps(component))
	instance, err := linker.Instantiate(store, component)
	require.NoError(t, err)
	require.NotNil(t, instance)
}

func TestComponentGetExportIndex(t *testing.T) {
	engine := newComponentEngine()

	wasm, err := Wat2Wasm(`
      (component
        (core module $m
          (func (export "hello"))
        )
        (core instance $i (instantiate $m))
        (func (export "hello") (canon lift (core func $i "hello")))
      )
    `)
	require.NoError(t, err)
	component, err := NewComponent(engine, wasm)
	require.NoError(t, err)
	defer component.Close()

	idx := component.GetExportIndex(nil, "hello")
	require.NotNil(t, idx, "expected to find hello export")
	defer idx.Close()

	missing := component.GetExportIndex(nil, "nope")
	require.Nil(t, missing, "expected nil for missing export")
}
