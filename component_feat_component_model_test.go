package wasmtime

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// helloComponent is the canonical "smallest non-trivial component" used by
// the Component-focused tests in this file: a single `hello` function with
// no arguments or results. Other test files define their own minimal
// fixtures with file-specific names rather than reaching across files for
// this constant.
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
	cases := []struct {
		name           string
		wat            string
		wantExportName string
	}{
		{
			"hello_export",
			`(component
  (core module $m (func (export "hello")))
  (core instance $i (instantiate $m))
  (func (export "hello") (canon lift (core func $i "hello"))))`,
			"hello",
		},
		{
			"world_export",
			`(component
  (core module $m (func (export "world")))
  (core instance $i (instantiate $m))
  (func (export "world") (canon lift (core func $i "world"))))`,
			"world",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			engine := newComponentEngine()
			component := newComponent(t, engine, tc.wat)
			defer component.Close()

			bytes, err := component.Serialize()
			require.NoError(t, err)
			require.NotEmpty(t, bytes)

			round, err := NewComponentDeserialize(engine, bytes)
			require.NoError(t, err)
			defer round.Close()
			require.NotNil(t, round.GetExportIndex(nil, tc.wantExportName))
		})
	}
}

func TestComponentSerializeRoundTripFile(t *testing.T) {
	cases := []struct {
		name           string
		wat            string
		wantExportName string
	}{
		{
			"hello_export",
			`(component
  (core module $m (func (export "hello")))
  (core instance $i (instantiate $m))
  (func (export "hello") (canon lift (core func $i "hello"))))`,
			"hello",
		},
		{
			"world_export",
			`(component
  (core module $m (func (export "world")))
  (core instance $i (instantiate $m))
  (func (export "world") (canon lift (core func $i "world"))))`,
			"world",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			engine := newComponentEngine()
			component := newComponent(t, engine, tc.wat)
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
			require.NotNil(t, round.GetExportIndex(nil, tc.wantExportName))
		})
	}
}

func TestComponentGetExportIndex(t *testing.T) {
	engine := newComponentEngine()
	component := newComponent(t, engine, helloComponent)
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
