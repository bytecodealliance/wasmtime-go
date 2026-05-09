package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// funcExportComponent is the canonical "minimal func export" fixture used
// by the type-walking tests below to exercise non-Type item kinds. It has
// a deliberately distinct name from helloComponent (in
// component_feat_component_model_test.go) so each test file owns its own
// fixtures rather than sharing a top-level constant across files.
const funcExportComponent = `
(component
  (core module $m (func (export "hello")))
  (core instance $i (instantiate $m))
  (func (export "hello") (canon lift (core func $i "hello"))))
`

// typeAliasU32Component exports a single WIT type alias `my-id = u32`.
// Used to observe the kind=Type vs kind!=Type axes (the alias name is
// also probed here so we know name handling works for the Type kind).
const typeAliasU32Component = `
(component
  (type $alias u32)
  (export "my-id" (type $alias)))
`

// missingHostInstanceComponent imports a single instance whose function
// will not be satisfied by an empty linker. Defined here for the
// type-test scenarios that walk the import side.
const missingHostInstanceComponent = `
(component
  (import "host:missing/api" (instance
    (export "f" (func)))))
`

func TestComponentTypeImportExportCount(t *testing.T) {
	cases := []struct {
		name        string
		wat         string
		wantImports int
		wantExports int
	}{
		{"single_export", funcExportComponent, 0, 1},
		{"single_import", missingHostInstanceComponent, 1, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			engine := newComponentEngine()
			component := newComponent(t, engine, tc.wat)
			defer component.Close()

			ct := component.Type()
			defer ct.Close()
			require.Equal(t, tc.wantImports, ct.ImportCount())
			require.Equal(t, tc.wantExports, ct.ExportCount())
		})
	}
}

func TestComponentTypeExportNthReturnsItem(t *testing.T) {
	engine := newComponentEngine()
	component := newComponent(t, engine, funcExportComponent)
	defer component.Close()

	ct := component.Type()
	defer ct.Close()

	name, item := ct.ExportNth(0)
	require.NotNil(t, item)
	defer item.Close()
	require.Equal(t, "hello", name)
	require.Equal(t, ComponentItemKindComponentFunc, item.Kind())
}

func TestComponentTypeExportNthOutOfRange(t *testing.T) {
	engine := newComponentEngine()
	component := newComponent(t, engine, funcExportComponent)
	defer component.Close()

	ct := component.Type()
	defer ct.Close()

	name, item := ct.ExportNth(1)
	require.Equal(t, "", name)
	require.Nil(t, item)
}

func TestComponentTypeImportNthReturnsItem(t *testing.T) {
	engine := newComponentEngine()
	component := newComponent(t, engine, missingHostInstanceComponent)
	defer component.Close()

	ct := component.Type()
	defer ct.Close()

	name, item := ct.ImportNth(0)
	require.NotNil(t, item)
	defer item.Close()
	require.Equal(t, "host:missing/api", name)
	require.Equal(t, ComponentItemKindComponentInstance, item.Kind())
}

func TestComponentTypeImportNthOutOfRange(t *testing.T) {
	engine := newComponentEngine()
	component := newComponent(t, engine, missingHostInstanceComponent)
	defer component.Close()

	ct := component.Type()
	defer ct.Close()

	name, item := ct.ImportNth(1)
	require.Equal(t, "", name)
	require.Nil(t, item)
}

func TestComponentItemTypeAlias(t *testing.T) {
	engine := newComponentEngine()
	component := newComponent(t, engine, typeAliasU32Component)
	defer component.Close()

	ct := component.Type()
	defer ct.Close()
	require.Equal(t, 1, ct.ExportCount())

	name, item := ct.ExportNth(0)
	require.NotNil(t, item)
	defer item.Close()
	require.Equal(t, "my-id", name)
	require.Equal(t, ComponentItemKindType, item.Kind())

	vt := item.TypeAlias()
	require.NotNil(t, vt)
	defer vt.Close()
	require.Equal(t, ComponentValTypeKindU32, vt.Kind())
}

func TestComponentItemTypeAliasReturnsNilForNonTypeKind(t *testing.T) {
	engine := newComponentEngine()
	component := newComponent(t, engine, funcExportComponent)
	defer component.Close()

	ct := component.Type()
	defer ct.Close()
	_, item := ct.ExportNth(0)
	require.NotNil(t, item)
	defer item.Close()
	// hello is a func export, not a type alias.
	require.Nil(t, item.TypeAlias())
}

// TestComponentTypeWalkMultipleTypeAliases verifies that ExportNth iterates
// correctly through a component carrying multiple primitive type aliases at
// distinct positions. Per-kind constant correctness is covered separately by
// TestComponentValTypeKindForEachPrimitive; this test focuses on the walking
// semantics — exercising ExportNth at indices > 0 and the (name, kind)
// coupling for each position in a single multi-type fixture.
func TestComponentTypeWalkMultipleTypeAliases(t *testing.T) {
	const wat = `(component
  (type $b bool)
  (type $i s32)
  (type $s string)
  (export "flag" (type $b))
  (export "count" (type $i))
  (export "label" (type $s)))`

	cases := []struct {
		index    int
		wantName string
		wantKind ComponentValTypeKind
	}{
		{0, "flag", ComponentValTypeKindBool},
		{1, "count", ComponentValTypeKindS32},
		{2, "label", ComponentValTypeKindString},
	}
	for _, tc := range cases {
		t.Run(tc.wantName, func(t *testing.T) {
			engine := newComponentEngine()
			component := newComponent(t, engine, wat)
			defer component.Close()

			ct := component.Type()
			defer ct.Close()
			require.Equal(t, 3, ct.ExportCount())

			name, item := ct.ExportNth(tc.index)
			require.NotNil(t, item)
			defer item.Close()
			require.Equal(t, tc.wantName, name)
			require.Equal(t, ComponentItemKindType, item.Kind())

			vt := item.TypeAlias()
			require.NotNil(t, vt)
			defer vt.Close()
			require.Equal(t, tc.wantKind, vt.Kind())
		})
	}
}

// TestComponentValTypeKindForEachPrimitive triangulates the
// ComponentValTypeKind* constants for the 13 primitive WIT types. Each
// case ships a minimal component containing a single type alias of one
// kind; the test walks the lone export and asserts its observed kind.
func TestComponentValTypeKindForEachPrimitive(t *testing.T) {
	cases := []struct {
		name     string
		wat      string
		wantKind ComponentValTypeKind
	}{
		{"bool", `(component (type $a bool) (export "a" (type $a)))`, ComponentValTypeKindBool},
		{"s8", `(component (type $a s8) (export "a" (type $a)))`, ComponentValTypeKindS8},
		{"s16", `(component (type $a s16) (export "a" (type $a)))`, ComponentValTypeKindS16},
		{"s32", `(component (type $a s32) (export "a" (type $a)))`, ComponentValTypeKindS32},
		{"s64", `(component (type $a s64) (export "a" (type $a)))`, ComponentValTypeKindS64},
		{"u8", `(component (type $a u8) (export "a" (type $a)))`, ComponentValTypeKindU8},
		{"u16", `(component (type $a u16) (export "a" (type $a)))`, ComponentValTypeKindU16},
		{"u32", `(component (type $a u32) (export "a" (type $a)))`, ComponentValTypeKindU32},
		{"u64", `(component (type $a u64) (export "a" (type $a)))`, ComponentValTypeKindU64},
		{"f32", `(component (type $a f32) (export "a" (type $a)))`, ComponentValTypeKindF32},
		{"f64", `(component (type $a f64) (export "a" (type $a)))`, ComponentValTypeKindF64},
		{"char", `(component (type $a char) (export "a" (type $a)))`, ComponentValTypeKindChar},
		{"string", `(component (type $a string) (export "a" (type $a)))`, ComponentValTypeKindString},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			engine := newComponentEngine()
			component := newComponent(t, engine, tc.wat)
			defer component.Close()

			ct := component.Type()
			defer ct.Close()

			_, item := ct.ExportNth(0)
			require.NotNil(t, item)
			defer item.Close()
			require.Equal(t, ComponentItemKindType, item.Kind())

			vt := item.TypeAlias()
			require.NotNil(t, vt)
			defer vt.Close()
			require.Equal(t, tc.wantKind, vt.Kind())
		})
	}
}
