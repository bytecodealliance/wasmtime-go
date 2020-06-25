package wasmtime

import "testing"

func TestTrap(t *testing.T) {
	trap := NewTrap(NewStore(NewEngine()), "message")
	if trap.Message() != "message" {
		panic("wrong message")
	}
}

func TestTrapFrames(t *testing.T) {
	store := NewStore(NewEngine())
	wasm, err := Wat2Wasm(`
	  (func call $foo)
	  (func $foo call $bar)
	  (func $bar unreachable)
	  (start 0)
	`)
	assertNoError(err)
	module, err := NewModule(store, wasm)
	assertNoError(err)

	i, err := NewInstance(store, module, []*Extern{})
	if i != nil {
		panic("expected failure")
	}
	if err == nil {
		panic("expected failure")
	}
	trap := err.(*Trap)
	frames := trap.Frames()
	if len(frames) != 3 {
		panic("expected 3 frames")
	}
	if *frames[0].FuncName() != "bar" {
		panic("bad function name")
	}
	if *frames[1].FuncName() != "foo" {
		panic("bad function name")
	}
	if frames[2].FuncName() != nil {
		panic("bad function name")
	}
	if frames[0].FuncIndex() != 2 {
		panic("bad function index")
	}
	if frames[1].FuncIndex() != 1 {
		panic("bad function index")
	}
	if frames[2].FuncIndex() != 0 {
		panic("bad function index")
	}

	expected := `wasm trap: unreachable
wasm backtrace:
  0:   0x26 - <unknown>!bar
  1:   0x21 - <unknown>!foo
  2:   0x1c - <unknown>!<wasm function 0>
`
	if trap.Error() != expected {
		t.Fatalf("expected\n%s\ngot\n%s", trap.Error(), expected)
	}
}

func TestTrapModuleName(t *testing.T) {
	store := NewStore(NewEngine())
	wasm, err := Wat2Wasm(`(module $f
	  (func unreachable)
	  (start 0)
	)`)
	assertNoError(err)
	module, err := NewModule(store, wasm)
	assertNoError(err)

	i, err := NewInstance(store, module, []*Extern{})
	if i != nil {
		panic("expected failure")
	}
	if err == nil {
		panic("expected failure")
	}
	trap := err.(*Trap)
	frames := trap.Frames()
	if len(frames) != 1 {
		panic("expected 3 frames")
	}
	if *frames[0].ModuleName() != "f" {
		panic("bad module name")
	}
}
