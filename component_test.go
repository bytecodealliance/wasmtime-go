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

// helloComponent is the smallest non-trivial component used by several tests:
// a single `hello` function with no arguments or results.
const helloComponent = `
(component
  (core module $m (func (export "hello")))
  (core instance $i (instantiate $m))
  (func (export "hello") (canon lift (core func $i "hello"))))
`

func TestComponentNew(t *testing.T) {
	helloWasm, err := Wat2Wasm(helloComponent)
	require.NoError(t, err)

	cases := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{"empty", []byte{}, true},
		{"garbage", []byte{1, 2, 3}, true},
		{"valid", helloWasm, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			engine := newComponentEngine()
			component, err := NewComponent(engine, tc.input)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, component)
			component.Close()
		})
	}
}

func TestComponentSerializeRoundTripBytes(t *testing.T) {
	engine := newComponentEngine()
	wasm, err := Wat2Wasm(helloComponent)
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
	require.NotNil(t, round.GetExportIndex(nil, "hello"))
}

func TestComponentSerializeRoundTripFile(t *testing.T) {
	engine := newComponentEngine()
	wasm, err := Wat2Wasm(helloComponent)
	require.NoError(t, err)
	component, err := NewComponent(engine, wasm)
	require.NoError(t, err)
	defer component.Close()

	bytes, err := component.Serialize()
	require.NoError(t, err)

	tmp, err := os.CreateTemp("", "component-serialize")
	require.NoError(t, err)
	defer os.Remove(tmp.Name())
	_, err = tmp.Write(bytes)
	require.NoError(t, err)
	require.NoError(t, tmp.Close())

	round, err := NewComponentDeserializeFile(engine, tmp.Name())
	require.NoError(t, err)
	defer round.Close()
	require.NotNil(t, round.GetExportIndex(nil, "hello"))
}

func TestComponentInstantiate(t *testing.T) {
	engine := newComponentEngine()
	store := NewStore(engine)

	wasm, err := Wat2Wasm(helloComponent)
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
	wasm, err := Wat2Wasm(helloComponent)
	require.NoError(t, err)
	component, err := NewComponent(engine, wasm)
	require.NoError(t, err)
	defer component.Close()

	cases := []struct {
		name      string
		target    string
		wantFound bool
	}{
		{"existing", "hello", true},
		{"missing", "nope", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			idx := component.GetExportIndex(nil, tc.target)
			if tc.wantFound {
				require.NotNil(t, idx)
				idx.Close()
			} else {
				require.Nil(t, idx)
			}
		})
	}
}
