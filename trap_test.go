package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrap(t *testing.T) {
	trap := NewTrap("message")
	require.Equal(t, trap.Message(), "message", "wrong message")
}

func TestTrapFrames(t *testing.T) {
	store := NewStore(NewEngine())
	wasm, err := Wat2Wasm(`
	  (func call $foo)
	  (func $foo call $bar)
	  (func $bar unreachable)
	  (start 0)
	`)
	require.NoError(t, err)
	module, err := NewModule(store.Engine, wasm)
	require.NoError(t, err)

	inst, err := NewInstance(store, module, []AsExtern{})
	require.Nil(t, inst, "expected failure")
	require.Error(t, err)

	trap := err.(*Trap)
	frames := trap.Frames()
	require.Len(t, frames, 3, "expected 3 frames")

	require.Equal(t, "bar", *frames[0].FuncName(), "bad function name")
	require.Equal(t, "foo", *frames[1].FuncName(), "bad function name")
	require.Nil(t, frames[2].FuncName(), "bad function name")
	require.Equal(t, uint32(2), frames[0].FuncIndex(), "bad function index")
	require.Equal(t, uint32(1), frames[1].FuncIndex(), "bad function index")
	require.Equal(t, uint32(0), frames[2].FuncIndex(), "bad function index")

	expected := `wasm trap: wasm ` + "`unreachable`" + ` instruction executed
wasm backtrace:
    0:   0x26 - <unknown>!bar
    1:   0x21 - <unknown>!foo
    2:   0x1c - <unknown>!<wasm function 0>
`

	require.Equal(t, expected, trap.Error())
	code := trap.Code()
	require.NotNil(t, code)
	require.Equal(t, *code, UnreachableCodeReached)
}

func TestTrapModuleName(t *testing.T) {
	store := NewStore(NewEngine())
	wasm, err := Wat2Wasm(`(module $f
	  (func unreachable)
	  (start 0)
	)`)
	require.NoError(t, err)
	module, err := NewModule(store.Engine, wasm)
	require.NoError(t, err)

	inst, err := NewInstance(store, module, []AsExtern{})
	require.Nil(t, inst, "expected failure")
	require.Error(t, err)

	trap := err.(*Trap)
	frames := trap.Frames()
	require.Len(t, frames, 1, "expected 1 frame")
	require.Equal(t, "f", *frames[0].ModuleName(), "bad function name")
}
