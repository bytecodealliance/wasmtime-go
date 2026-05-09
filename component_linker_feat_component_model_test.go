package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComponentInstantiate(t *testing.T) {
	cases := []struct {
		name string
		wat  string
	}{
		{
			"hello_component",
			`(component
  (core module $m (func (export "hello")))
  (core instance $i (instantiate $m))
  (func (export "hello") (canon lift (core func $i "hello"))))`,
		},
		{
			"empty_component",
			`(component)`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			engine := newComponentEngine()
			store := NewStore(engine)
			component := newComponent(t, engine, tc.wat)
			defer component.Close()

			linker := NewComponentLinker(engine)
			defer linker.Close()

			instance, err := linker.Instantiate(store, component)
			require.NoError(t, err)
			require.NotNil(t, instance)
		})
	}
}

func TestComponentDefineUnknownImportsAsTraps(t *testing.T) {
	cases := []struct {
		name string
		wat  string
	}{
		{
			"single_missing_instance",
			`(component
  (import "host:missing/api" (instance
    (export "f" (func)))))`,
		},
		{
			"two_missing_instances",
			`(component
  (import "host:a/api" (instance
    (export "f" (func))))
  (import "host:b/api" (instance
    (export "g" (func)))))`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			engine := newComponentEngine()
			store := NewStore(engine)
			component := newComponent(t, engine, tc.wat)
			defer component.Close()

			linker := NewComponentLinker(engine)
			defer linker.Close()

			// Without trap definitions the missing imports cause instantiation to fail.
			_, err := linker.Instantiate(store, component)
			require.Error(t, err, "expected missing import to fail instantiation")

			// After defining unknown imports as traps the instantiation succeeds.
			require.NoError(t, linker.DefineUnknownImportsAsTraps(component))
			instance, err := linker.Instantiate(store, component)
			require.NoError(t, err)
			require.NotNil(t, instance)
		})
	}
}
