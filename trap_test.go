package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrap(t *testing.T) {
	trap := NewTrap("message")
	assert.Equal(t, trap.Message(), "message", "wrong message")
}

func TestTrapFrames(t *testing.T) {
	store := NewStore(NewEngine())
	wasm, err := Wat2Wasm(`
	  (func call $foo)
	  (func $foo call $bar)
	  (func $bar unreachable)
	  (start 0)
	`)
	assert.NoError(t, err)
	module, err := NewModule(store.Engine, wasm)
	assert.NoError(t, err)

	inst, err := NewInstance(store, module, []AsExtern{})
	assert.Nil(t, inst, "expected failure")
	assert.Error(t, err)

	trap := err.(*Trap)
	frames := trap.Frames()
	assert.Len(t, frames, 3, "expected 3 frames")

	assert.Equal(t, "bar", *frames[0].FuncName(), "bad function name")
	assert.Equal(t, "foo", *frames[1].FuncName(), "bad function name")
	assert.Equal(t, nil, frames[2].FuncName(), "bad function name")
	assert.Equal(t, 2, frames[0].FuncIndex(), "bad function index")
	assert.Equal(t, 1, frames[1].FuncIndex(), "bad function index")
	assert.Equal(t, 0, frames[2].FuncIndex(), "bad function index")

	expected := `wasm trap: wasm ` + "`unreachable`" + ` instruction executed
wasm backtrace:
    0:   0x26 - <unknown>!bar
    1:   0x21 - <unknown>!foo
    2:   0x1c - <unknown>!<wasm function 0>
`

	assert.Equal(t, expected, trap.Error())
	code := trap.Code()
	assert.NotNil(t, code)
	assert.Equal(t, *code, UnreachableCodeReached)
}

func TestTrapModuleName(t *testing.T) {
	store := NewStore(NewEngine())
	wasm, err := Wat2Wasm(`(module $f
	  (func unreachable)
	  (start 0)
	)`)
	assert.NoError(t, err)
	module, err := NewModule(store.Engine, wasm)
	assert.NoError(t, err)

	inst, err := NewInstance(store, module, []AsExtern{})
	assert.Nil(t, inst, "expected failure")
	assert.Error(t, err)

	trap := err.(*Trap)
	frames := trap.Frames()
	assert.Len(t, frames, 1, "expected 1 frame")
	assert.Equal(t, "f", *frames[0].FuncName(), "bad function name")
}
