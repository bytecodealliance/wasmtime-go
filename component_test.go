package wasmtime

// This file holds test helpers and fixture WATs that are shared by the
// component-model test files (component_feat_component_model_test.go,
// component_linker_feat_component_model_test.go,
// component_type_feat_component_model_test.go). It does not correspond to
// a single production source file; the 1:1 production-test correspondence
// still applies to the test functions in the other files.

import (
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

// newComponent compiles `wat` into a fresh [Component] for tests. The
// returned component must be closed by the caller.
func newComponent(t *testing.T, engine *Engine, wat string) *Component {
	t.Helper()
	wasm, err := Wat2Wasm(wat)
	require.NoError(t, err)
	component, err := NewComponent(engine, wasm)
	require.NoError(t, err)
	return component
}
