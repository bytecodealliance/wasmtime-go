//go:build wasmtime_component_model

package wasmtime

import (
	"reflect"
	"testing"
)

const byteComponent = `(component
  (core module $m
    (memory (export "memory") 1)
    (global $next (mut i32) (i32.const 1024))
    (func (export "realloc")
      (param $old i32) (param $old-size i32) (param $align i32) (param $new-size i32)
      (result i32)
      (local $ptr i32)
      (local.set $ptr (global.get $next))
      (global.set $next (i32.add (local.get $ptr) (local.get $new-size)))
      (local.get $ptr))
    (func (export "increment") (param $ptr i32) (param $len i32) (result i32)
      (local $i i32)
      (loop $loop
        (if (i32.lt_u (local.get $i) (local.get $len))
          (then
            (i32.store8
              (i32.add (local.get $ptr) (local.get $i))
              (i32.add
                (i32.load8_u (i32.add (local.get $ptr) (local.get $i)))
                (i32.const 1)))
            (local.set $i (i32.add (local.get $i) (i32.const 1)))
            (br $loop))))
      (i32.store (i32.const 0) (local.get $ptr))
      (i32.store offset=4 (i32.const 0) (local.get $len))
      (i32.const 0)))
  (core instance $i (instantiate $m))
  (func (export "increment-bytes") (param "bytes" (list u8)) (result (list u8))
    (canon lift (core func $i "increment")
      (memory $i "memory")
      (realloc (func $i "realloc")))))`

func closeComponentVals(values []*ComponentVal) {
	for _, value := range values {
		value.Close()
	}
}

func TestComponentFuncRepeatedListU8Calls(t *testing.T) {
	engine := newComponentEngine()
	defer engine.Close()
	component := newComponent(t, engine, byteComponent)
	defer component.Close()
	store := NewStore(engine)
	linker := NewComponentLinker(engine)
	defer linker.Close()
	instance, err := linker.Instantiate(store, component)
	if err != nil {
		t.Fatal(err)
	}
	increment := instance.GetFunc(store, nil, "increment-bytes")
	if increment == nil {
		t.Fatal("increment-bytes function not found")
	}
	value := []byte{0, 1, 2, 253}
	for iteration := 0; iteration < 100; iteration++ {
		elements := make([]*ComponentVal, len(value))
		for i, item := range value {
			elements[i] = NewComponentU8(item)
		}
		argument := NewComponentList(elements)
		closeComponentVals(elements)
		results, err := increment.Call(store, []*ComponentVal{argument})
		argument.Close()
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 1 || results[0].Kind() != ComponentValKindList {
			t.Fatalf("unexpected results: %#v", results)
		}
		returned := results[0].Value().([]*ComponentVal)
		value = value[:0]
		for _, item := range returned {
			value = append(value, item.Value().(uint8))
		}
		closeComponentVals(returned)
		closeComponentVals(results)
	}
	if want := []byte{100, 101, 102, 97}; !reflect.DeepEqual(value, want) {
		t.Fatalf("value = %v, want %v", value, want)
	}
}

func TestComponentValCompositeOwnership(t *testing.T) {
	primitives := []struct {
		value *ComponentVal
		want  any
	}{
		{NewComponentBool(true), true},
		{NewComponentS8(-8), int8(-8)},
		{NewComponentU8(8), uint8(8)},
		{NewComponentS16(-16), int16(-16)},
		{NewComponentU16(16), uint16(16)},
		{NewComponentS32(-32), int32(-32)},
		{NewComponentU32(32), uint32(32)},
		{NewComponentS64(-64), int64(-64)},
		{NewComponentU64(64), uint64(64)},
		{NewComponentF32(3.25), float32(3.25)},
		{NewComponentF64(6.5), float64(6.5)},
		{NewComponentChar('F'), rune('F')},
		{NewComponentString("Friday"), "Friday"},
	}
	for _, primitive := range primitives {
		if got := primitive.value.Value(); got != primitive.want {
			t.Errorf("primitive value = %#v, want %#v", got, primitive.want)
		}
		primitive.value.Close()
	}

	left := NewComponentS32(-7)
	right := NewComponentU32(49)
	record := NewComponentRecord([]ComponentRecordField{{Name: "left", Value: left}, {Name: "right", Value: right}})
	left.Close()
	right.Close()
	recordFields := record.Value().([]ComponentRecordField)
	if recordFields[0].Name != "left" || recordFields[0].Value.Value() != int32(-7) || recordFields[1].Name != "right" || recordFields[1].Value.Value() != uint32(49) {
		t.Fatalf("unexpected record: %#v", recordFields)
	}
	closeComponentVals([]*ComponentVal{recordFields[0].Value, recordFields[1].Value})

	payload := NewComponentString("payload")
	values := []*ComponentVal{
		NewComponentTuple([]*ComponentVal{payload}),
		NewComponentVariant("case", payload),
		NewComponentOption(payload),
		NewComponentOption(nil),
		NewComponentResult(true, payload),
		NewComponentResult(false, nil),
		NewComponentEnum("choice"),
		NewComponentFlags([]string{"read", "write"}),
	}
	payload.Close()

	tuple := values[0].Value().([]*ComponentVal)
	if tuple[0].Value() != "payload" {
		t.Fatalf("unexpected tuple: %#v", tuple)
	}
	closeComponentVals(tuple)
	variant := values[1].Value().(ComponentVariantValue)
	if variant.Discriminant != "case" || variant.Value.Value() != "payload" {
		t.Fatalf("unexpected variant: %#v", variant)
	}
	variant.Value.Close()
	option := values[2].Value().(*ComponentVal)
	if option.Value() != "payload" {
		t.Fatalf("unexpected option: %#v", option)
	}
	option.Close()
	if values[3].Value().(*ComponentVal) != nil {
		t.Fatal("none option returned a payload")
	}
	ok := values[4].Value().(ComponentResultValue)
	if !ok.OK || ok.Value.Value() != "payload" {
		t.Fatalf("unexpected ok result: %#v", ok)
	}
	ok.Value.Close()
	errResult := values[5].Value().(ComponentResultValue)
	if errResult.OK || errResult.Value != nil {
		t.Fatalf("unexpected error result: %#v", errResult)
	}
	if values[6].Value() != "choice" || !reflect.DeepEqual(values[7].Value(), []string{"read", "write"}) {
		t.Fatal("enum or flags did not round trip")
	}
	clone := record.Clone()
	record.Close()
	if clone.Kind() != ComponentValKindRecord {
		t.Fatal("deep clone did not survive source close")
	}
	clone.Close()
	closeComponentVals(values)
}
