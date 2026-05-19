package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComponentValTypeList(t *testing.T) {
	engine := newComponentEngine()
	wasm, err := Wat2Wasm(`(component
  (type $l (list u32))
  (export "l" (type $l)))`)
	require.NoError(t, err)
	component, err := NewComponent(engine, wasm)
	require.NoError(t, err)
	defer component.Close()

	ct := component.Type()
	defer ct.Close()
	_, item := ct.ExportNth(0)
	require.NotNil(t, item)
	defer item.Close()

	vt := item.TypeAlias()
	require.NotNil(t, vt)
	defer vt.Close()
	require.Equal(t, ComponentValTypeKindList, vt.Kind())

	lt := vt.List()
	require.NotNil(t, lt)
	defer lt.Close()

	elem := lt.Element()
	require.NotNil(t, elem)
	defer elem.Close()
	require.Equal(t, ComponentValTypeKindU32, elem.Kind())
}

func TestComponentValTypeRecord(t *testing.T) {
	engine := newComponentEngine()
	wasm, err := Wat2Wasm(`(component
  (type $r (record (field "age" u32) (field "name" string)))
  (export "r" (type $r)))`)
	require.NoError(t, err)
	component, err := NewComponent(engine, wasm)
	require.NoError(t, err)
	defer component.Close()

	ct := component.Type()
	defer ct.Close()
	_, item := ct.ExportNth(0)
	defer item.Close()
	vt := item.TypeAlias()
	defer vt.Close()
	require.Equal(t, ComponentValTypeKindRecord, vt.Kind())

	rt := vt.Record()
	require.NotNil(t, rt)
	defer rt.Close()
	require.Equal(t, 2, rt.FieldCount())

	name0, ty0 := rt.FieldNth(0)
	require.Equal(t, "age", name0)
	require.NotNil(t, ty0)
	defer ty0.Close()
	require.Equal(t, ComponentValTypeKindU32, ty0.Kind())

	name1, ty1 := rt.FieldNth(1)
	require.Equal(t, "name", name1)
	require.NotNil(t, ty1)
	defer ty1.Close()
	require.Equal(t, ComponentValTypeKindString, ty1.Kind())

	// FieldNth out of range returns ("", nil).
	name2, ty2 := rt.FieldNth(2)
	require.Equal(t, "", name2)
	require.Nil(t, ty2)
}

func TestComponentValTypeTuple(t *testing.T) {
	engine := newComponentEngine()
	wasm, err := Wat2Wasm(`(component
  (type $t (tuple bool s32 string))
  (export "t" (type $t)))`)
	require.NoError(t, err)
	component, err := NewComponent(engine, wasm)
	require.NoError(t, err)
	defer component.Close()

	ct := component.Type()
	defer ct.Close()
	_, item := ct.ExportNth(0)
	defer item.Close()
	vt := item.TypeAlias()
	defer vt.Close()
	require.Equal(t, ComponentValTypeKindTuple, vt.Kind())

	tt := vt.Tuple()
	require.NotNil(t, tt)
	defer tt.Close()
	require.Equal(t, 3, tt.TypeCount())

	ty0 := tt.TypeNth(0)
	require.NotNil(t, ty0)
	defer ty0.Close()
	require.Equal(t, ComponentValTypeKindBool, ty0.Kind())

	ty1 := tt.TypeNth(1)
	require.NotNil(t, ty1)
	defer ty1.Close()
	require.Equal(t, ComponentValTypeKindS32, ty1.Kind())

	ty2 := tt.TypeNth(2)
	require.NotNil(t, ty2)
	defer ty2.Close()
	require.Equal(t, ComponentValTypeKindString, ty2.Kind())

	// TypeNth out of range returns nil.
	require.Nil(t, tt.TypeNth(3))
}

func TestComponentValTypeEnum(t *testing.T) {
	engine := newComponentEngine()
	wasm, err := Wat2Wasm(`(component
  (type $e (enum "red" "green" "blue"))
  (export "e" (type $e)))`)
	require.NoError(t, err)
	component, err := NewComponent(engine, wasm)
	require.NoError(t, err)
	defer component.Close()

	ct := component.Type()
	defer ct.Close()
	_, item := ct.ExportNth(0)
	defer item.Close()
	vt := item.TypeAlias()
	defer vt.Close()
	require.Equal(t, ComponentValTypeKindEnum, vt.Kind())

	et := vt.Enum()
	require.NotNil(t, et)
	defer et.Close()
	require.Equal(t, 3, et.CaseCount())
	require.Equal(t, "red", et.CaseNth(0))
	require.Equal(t, "green", et.CaseNth(1))
	require.Equal(t, "blue", et.CaseNth(2))

	// CaseNth out of range returns "".
	require.Equal(t, "", et.CaseNth(3))
}

func TestComponentValTypeFlags(t *testing.T) {
	engine := newComponentEngine()
	wasm, err := Wat2Wasm(`(component
  (type $f (flags "read" "write" "exec"))
  (export "f" (type $f)))`)
	require.NoError(t, err)
	component, err := NewComponent(engine, wasm)
	require.NoError(t, err)
	defer component.Close()

	ct := component.Type()
	defer ct.Close()
	_, item := ct.ExportNth(0)
	defer item.Close()
	vt := item.TypeAlias()
	defer vt.Close()
	require.Equal(t, ComponentValTypeKindFlags, vt.Kind())

	ft := vt.Flags()
	require.NotNil(t, ft)
	defer ft.Close()
	require.Equal(t, 3, ft.FlagCount())
	require.Equal(t, "read", ft.FlagNth(0))
	require.Equal(t, "write", ft.FlagNth(1))
	require.Equal(t, "exec", ft.FlagNth(2))

	// FlagNth out of range returns "".
	require.Equal(t, "", ft.FlagNth(3))
}

// TestComponentValTypeDowncastNilForOtherKinds checks that each composite
// downcast method ([ComponentValType.List], [ComponentValType.Record],
// [ComponentValType.Tuple], [ComponentValType.Enum],
// [ComponentValType.Flags]) returns nil when invoked on a value type of an
// unrelated kind. A single `u32` type alias serves as the unrelated kind
// for all five probes.
func TestComponentValTypeDowncastNilForOtherKinds(t *testing.T) {
	engine := newComponentEngine()
	wasm, err := Wat2Wasm(`(component (type $a u32) (export "a" (type $a)))`)
	require.NoError(t, err)
	component, err := NewComponent(engine, wasm)
	require.NoError(t, err)
	defer component.Close()

	ct := component.Type()
	defer ct.Close()
	_, item := ct.ExportNth(0)
	defer item.Close()
	vt := item.TypeAlias()
	defer vt.Close()
	require.Equal(t, ComponentValTypeKindU32, vt.Kind())

	require.Nil(t, vt.List())
	require.Nil(t, vt.Record())
	require.Nil(t, vt.Tuple())
	require.Nil(t, vt.Enum())
	require.Nil(t, vt.Flags())
}
