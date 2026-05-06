package wasmtime

import (
	"fmt"
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

// primitiveTestComponent is a component that exports identity functions for
// every primitive WIT type the Go bindings currently support. The four core
// identity functions (one per core wasm value type) are reused via different
// component-level type signatures, since the canonical ABI takes care of
// converting between WIT types and their core representations.
const primitiveTestComponent = `
(component
  (core module $m
    (func (export "id_i32") (param i32) (result i32) local.get 0)
    (func (export "id_i64") (param i64) (result i64) local.get 0)
    (func (export "id-f32") (param f32) (result f32) local.get 0)
    (func (export "id-f64") (param f64) (result f64) local.get 0)
    (func (export "first_i32") (param i32 i32) (result i32) local.get 0)
    (func (export "noop")))
  (core instance $i (instantiate $m))
  (func (export "noop") (canon lift (core func $i "noop")))
  (func (export "id-bool") (param "x" bool) (result bool) (canon lift (core func $i "id_i32")))
  (func (export "id-s8")   (param "x" s8)   (result s8)   (canon lift (core func $i "id_i32")))
  (func (export "id-u8")   (param "x" u8)   (result u8)   (canon lift (core func $i "id_i32")))
  (func (export "id-s16")  (param "x" s16)  (result s16)  (canon lift (core func $i "id_i32")))
  (func (export "id-u16")  (param "x" u16)  (result u16)  (canon lift (core func $i "id_i32")))
  (func (export "id-s32")  (param "x" s32)  (result s32)  (canon lift (core func $i "id_i32")))
  (func (export "id-u32")  (param "x" u32)  (result u32)  (canon lift (core func $i "id_i32")))
  (func (export "id-s64")  (param "x" s64)  (result s64)  (canon lift (core func $i "id_i64")))
  (func (export "id-u64")  (param "x" u64)  (result u64)  (canon lift (core func $i "id_i64")))
  (func (export "id-f32")  (param "x" f32)  (result f32)  (canon lift (core func $i "id-f32")))
  (func (export "id-f64")  (param "x" f64)  (result f64)  (canon lift (core func $i "id-f64")))
  (func (export "id-char") (param "c" char) (result char) (canon lift (core func $i "id_i32")))
  (func (export "first-s32") (param "a" s32) (param "b" s32) (result s32) (canon lift (core func $i "first_i32"))))
`

func setupPrimitiveTest(t *testing.T) (*Store, *ComponentInstance) {
	t.Helper()
	engine := newComponentEngine()
	store := NewStore(engine)

	wasm, err := Wat2Wasm(primitiveTestComponent)
	require.NoError(t, err)

	component, err := NewComponent(engine, wasm)
	require.NoError(t, err)
	t.Cleanup(component.Close)

	linker := NewComponentLinker(engine)
	t.Cleanup(linker.Close)

	instance, err := linker.Instantiate(store, component)
	require.NoError(t, err)

	return store, instance
}

func TestComponentInstanceGetFunc(t *testing.T) {
	store, instance := setupPrimitiveTest(t)
	require.NotNil(t, instance.GetFunc(store, "id-s32"))
	require.Nil(t, instance.GetFunc(store, "nope"))
}

func TestComponentCallVoid(t *testing.T) {
	store, instance := setupPrimitiveTest(t)
	f := instance.GetFunc(store, "noop")
	require.NotNil(t, f)
	got, err := f.Call(store)
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestComponentCallBool(t *testing.T) {
	store, instance := setupPrimitiveTest(t)
	f := instance.GetFunc(store, "id-bool")
	require.NotNil(t, f)
	for _, want := range []bool{true, false} {
		got, err := f.Call(store, want)
		require.NoError(t, err)
		require.Equal(t, want, got)
	}
}

func TestComponentCallSignedIntegers(t *testing.T) {
	store, instance := setupPrimitiveTest(t)
	cases := []struct {
		name string
		val  interface{}
	}{
		{"id-s8", int8(-128)},
		{"id-s8", int8(127)},
		{"id-s16", int16(-32768)},
		{"id-s16", int16(32767)},
		{"id-s32", int32(-2147483648)},
		{"id-s32", int32(2147483647)},
		{"id-s64", int64(-9223372036854775808)},
		{"id-s64", int64(9223372036854775807)},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s/%v", tc.name, tc.val), func(t *testing.T) {
			f := instance.GetFunc(store, tc.name)
			require.NotNil(t, f)
			got, err := f.Call(store, tc.val)
			require.NoError(t, err)
			require.Equal(t, tc.val, got)
		})
	}
}

func TestComponentCallUnsignedIntegers(t *testing.T) {
	store, instance := setupPrimitiveTest(t)
	cases := []struct {
		name string
		val  interface{}
	}{
		{"id-u8", uint8(0)},
		{"id-u8", uint8(255)},
		{"id-u16", uint16(0)},
		{"id-u16", uint16(65535)},
		{"id-u32", uint32(0)},
		{"id-u32", uint32(4294967295)},
		{"id-u64", uint64(0)},
		{"id-u64", uint64(18446744073709551615)},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s/%v", tc.name, tc.val), func(t *testing.T) {
			f := instance.GetFunc(store, tc.name)
			require.NotNil(t, f)
			got, err := f.Call(store, tc.val)
			require.NoError(t, err)
			require.Equal(t, tc.val, got)
		})
	}
}

func TestComponentCallFloats(t *testing.T) {
	store, instance := setupPrimitiveTest(t)

	f32fn := instance.GetFunc(store, "id-f32")
	require.NotNil(t, f32fn)
	got, err := f32fn.Call(store, float32(3.14))
	require.NoError(t, err)
	require.Equal(t, float32(3.14), got)

	f64fn := instance.GetFunc(store, "id-f64")
	require.NotNil(t, f64fn)
	got, err = f64fn.Call(store, float64(2.718281828))
	require.NoError(t, err)
	require.Equal(t, float64(2.718281828), got)
}

func TestComponentCallChar(t *testing.T) {
	store, instance := setupPrimitiveTest(t)
	f := instance.GetFunc(store, "id-char")
	require.NotNil(t, f)

	// rune (= int32) is the natural type for char
	got, err := f.Call(store, rune('A'))
	require.NoError(t, err)
	require.Equal(t, rune('A'), got)

	// Multi-byte codepoint round-trip
	got, err = f.Call(store, rune('あ')) // Hiragana 'a'
	require.NoError(t, err)
	require.Equal(t, rune('あ'), got)
}

func TestComponentCallMultipleParams(t *testing.T) {
	store, instance := setupPrimitiveTest(t)
	f := instance.GetFunc(store, "first-s32")
	require.NotNil(t, f)
	// `first_s32` is wired to a core function that returns its first argument,
	// so this verifies that multiple arguments are marshaled in the right
	// order.
	got, err := f.Call(store, int32(7), int32(35))
	require.NoError(t, err)
	require.Equal(t, int32(7), got)
}

func TestComponentCallWrongArgCount(t *testing.T) {
	store, instance := setupPrimitiveTest(t)
	f := instance.GetFunc(store, "id-s32")
	require.NotNil(t, f)

	_, err := f.Call(store)
	require.Error(t, err)
	_, err = f.Call(store, int32(1), int32(2))
	require.Error(t, err)
}

func TestComponentCallWrongArgType(t *testing.T) {
	store, instance := setupPrimitiveTest(t)
	f := instance.GetFunc(store, "id-s32")
	require.NotNil(t, f)
	_, err := f.Call(store, int64(1)) // expects int32
	require.Error(t, err)
}
