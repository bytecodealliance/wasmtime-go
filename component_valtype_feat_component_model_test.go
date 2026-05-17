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

func TestComponentValTypeListOnNonListReturnsNil(t *testing.T) {
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
	require.Nil(t, vt.List())
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
	require.NotNil(t, item)
	defer item.Close()

	vt := item.TypeAlias()
	require.NotNil(t, vt)
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
}

func TestComponentValTypeRecordFieldNthOutOfRange(t *testing.T) {
	engine := newComponentEngine()
	wasm, err := Wat2Wasm(`(component
  (type $r (record (field "x" u32)))
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
	rt := vt.Record()
	defer rt.Close()

	name, ty := rt.FieldNth(1)
	require.Equal(t, "", name)
	require.Nil(t, ty)
}

func TestComponentValTypeRecordOnNonRecordReturnsNil(t *testing.T) {
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
	require.Nil(t, vt.Record())
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
	require.Equal(t, 3, tt.TypesCount())

	ty0 := tt.TypesNth(0)
	require.NotNil(t, ty0)
	defer ty0.Close()
	require.Equal(t, ComponentValTypeKindBool, ty0.Kind())

	ty1 := tt.TypesNth(1)
	require.NotNil(t, ty1)
	defer ty1.Close()
	require.Equal(t, ComponentValTypeKindS32, ty1.Kind())

	ty2 := tt.TypesNth(2)
	require.NotNil(t, ty2)
	defer ty2.Close()
	require.Equal(t, ComponentValTypeKindString, ty2.Kind())
}

func TestComponentValTypeTupleTypesNthOutOfRange(t *testing.T) {
	engine := newComponentEngine()
	wasm, err := Wat2Wasm(`(component
  (type $t (tuple bool))
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
	tt := vt.Tuple()
	defer tt.Close()

	require.Nil(t, tt.TypesNth(1))
}

func TestComponentValTypeTupleOnNonTupleReturnsNil(t *testing.T) {
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
	require.Nil(t, vt.Tuple())
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
	require.Equal(t, 3, et.NamesCount())
	require.Equal(t, "red", et.NamesNth(0))
	require.Equal(t, "green", et.NamesNth(1))
	require.Equal(t, "blue", et.NamesNth(2))
}

func TestComponentValTypeEnumNamesNthOutOfRange(t *testing.T) {
	engine := newComponentEngine()
	wasm, err := Wat2Wasm(`(component
  (type $e (enum "only"))
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
	et := vt.Enum()
	defer et.Close()

	require.Equal(t, "", et.NamesNth(1))
}

func TestComponentValTypeEnumOnNonEnumReturnsNil(t *testing.T) {
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
	require.Nil(t, vt.Enum())
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
	require.Equal(t, 3, ft.NamesCount())
	require.Equal(t, "read", ft.NamesNth(0))
	require.Equal(t, "write", ft.NamesNth(1))
	require.Equal(t, "exec", ft.NamesNth(2))
}

func TestComponentValTypeFlagsNamesNthOutOfRange(t *testing.T) {
	engine := newComponentEngine()
	wasm, err := Wat2Wasm(`(component
  (type $f (flags "only"))
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
	ft := vt.Flags()
	defer ft.Close()

	require.Equal(t, "", ft.NamesNth(1))
}

func TestComponentValTypeFlagsOnNonFlagsReturnsNil(t *testing.T) {
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
	require.Nil(t, vt.Flags())
}
