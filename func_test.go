package wasmtime

import (
	"fmt"
	"strings"
	"testing"
)

func TestFunc(t *testing.T) {
	store := NewStore(NewEngine())
	cb := func(caller *Caller, args []Val) ([]Val, *Trap) {
		return []Val{}, nil
	}
	NewFunc(store, NewFuncType([]*ValType{}, []*ValType{}), cb)
}

func TestFuncCall(t *testing.T) {
	store := NewStore(NewEngine())
	called := false
	cb := func(caller *Caller, args []Val) ([]Val, *Trap) {
		called = true
		return []Val{}, nil
	}
	f := NewFunc(store, NewFuncType([]*ValType{}, []*ValType{}), cb)
	results, trap := f.Call(store)
	if trap != nil {
		panic(trap)
	}
	if results != nil {
		panic("bad results")
	}
	if !called {
		panic("didn't call")
	}
}

func TestFuncTrap(t *testing.T) {
	store := NewStore(NewEngine())
	cb := func(caller *Caller, args []Val) ([]Val, *Trap) {
		return nil, NewTrap("x")
	}
	f := NewFunc(store, NewFuncType([]*ValType{}, []*ValType{}), cb)
	results, err := f.Call(store)
	if err == nil {
		panic("bad trap")
	}
	if results != nil {
		panic("bad results")
	}
	trap := err.(*Trap)
	if trap.Message() != "x" {
		panic("bad message")
	}
}

func TestFuncPanic(t *testing.T) {
	store := NewStore(NewEngine())
	cb := func(caller *Caller, args []Val) ([]Val, *Trap) {
		panic("x")
	}
	f := NewFunc(store, NewFuncType([]*ValType{}, []*ValType{}), cb)
	var caught interface{}
	var results interface{}
	var err error
	func() {
		defer func() { caught = recover() }()
		results, err = f.Call(store)
	}()
	if caught == nil {
		panic("panic didn't work")
	}
	if caught.(string) != "x" {
		panic("value didn't propagate")
	}
	if err != nil {
		panic("bad trap")
	}
	if results != nil {
		panic("bad results")
	}
}

func TestFuncArgs(t *testing.T) {
	store := NewStore(NewEngine())
	cb := func(caller *Caller, args []Val) ([]Val, *Trap) {
		if len(args) != 2 {
			panic("wrong argument size")
		}
		if args[0].I32() != 1 {
			panic("bad first argument")
		}
		if args[1].I64() != 2 {
			panic("bad second argument")
		}
		return []Val{ValF32(3), ValF64(4)}, nil
	}
	i32 := NewValType(KindI32)
	i64 := NewValType(KindI64)
	f32 := NewValType(KindF32)
	f64 := NewValType(KindF64)
	f := NewFunc(store, NewFuncType([]*ValType{i32, i64}, []*ValType{f32, f64}), cb)
	results, trap := f.Call(store, int32(1), int64(2))
	if trap != nil {
		panic(trap)
	}
	list := results.([]Val)
	if len(list) != 2 {
		panic("bad results")
	}
	if list[0].F32() != 3 {
		panic("bad result[0]")
	}
	if list[1].F64() != 4 {
		panic("bad result[1]")
	}
}

func TestFuncOneRet(t *testing.T) {
	store := NewStore(NewEngine())
	cb := func(caller *Caller, args []Val) ([]Val, *Trap) {
		if len(args) != 0 {
			panic("wrong argument size")
		}
		return []Val{ValI32(3)}, nil
	}
	i32 := NewValType(KindI32)
	f := NewFunc(store, NewFuncType([]*ValType{}, []*ValType{i32}), cb)
	results, trap := f.Call(store)
	if trap != nil {
		panic(trap)
	}
	result := results.(int32)
	if result != 3 {
		panic("bad result")
	}
}

func TestFuncWrongRet(t *testing.T) {
	store := NewStore(NewEngine())
	cb := func(caller *Caller, args []Val) ([]Val, *Trap) {
		return []Val{ValI64(3)}, nil
	}
	i32 := NewValType(KindI32)
	f := NewFunc(store, NewFuncType([]*ValType{}, []*ValType{i32}), cb)

	var caught interface{}
	func() {
		defer func() { caught = recover() }()
		f.Call(store)
	}()
	if caught == nil {
		panic("expected a panic")
	}
	if !strings.Contains(caught.(string), "callback produced wrong type of result") {
		panic(fmt.Sprintf("wrong panic message %s", caught))
	}
}

func TestFuncWrongRet2(t *testing.T) {
	store := NewStore(NewEngine())
	cb := func(caller *Caller, args []Val) ([]Val, *Trap) {
		return []Val{}, nil
	}
	i32 := NewValType(KindI32)
	f := NewFunc(store, NewFuncType([]*ValType{}, []*ValType{i32}), cb)

	var caught interface{}
	func() {
		defer func() { caught = recover() }()
		f.Call(store)
	}()
	if caught == nil {
		panic("expected a panic")
	}
	if !strings.Contains(caught.(string), "callback didn't produce the correct number of results") {
		panic(fmt.Sprintf("wrong panic message %s", caught))
	}
}

func TestFuncWrapSimple(t *testing.T) {
	store := NewStore(NewEngine())
	called := false
	f := WrapFunc(store, func() {
		called = true
	})
	result, trap := f.Call(store)
	if trap != nil {
		panic(trap)
	}
	if result != nil {
		panic("wrong result")
	}
	if !called {
		panic("not called")
	}
}

func TestFuncWrapSimple1Arg(t *testing.T) {
	store := NewStore(NewEngine())
	f := WrapFunc(store, func(i int32) {
		if i != 3 {
			panic("wrong argument")
		}
	})
	result, trap := f.Call(store, 3)
	if trap != nil {
		panic(trap)
	}
	if result != nil {
		panic("wrong result")
	}
}

func TestFuncWrapSimpleManyArg(t *testing.T) {
	store := NewStore(NewEngine())
	f := WrapFunc(store, func(i1 int32, i2 int64, f1 float32, f2 float64) {
		if i1 != 3 {
			panic("wrong argument")
		}
		if i2 != 4 {
			panic("wrong argument")
		}
		if f1 != 5 {
			panic("wrong argument")
		}
		if f2 != 6 {
			panic("wrong argument")
		}
	})
	result, trap := f.Call(store, 3, 4, float32(5.0), float64(6.0))
	if trap != nil {
		panic(trap)
	}
	if result != nil {
		panic("wrong result")
	}
}

func TestFuncWrapCallerArg(t *testing.T) {
	store := NewStore(NewEngine())
	f := WrapFunc(store, func(c *Caller) {})
	result, trap := f.Call(store)
	if trap != nil {
		panic(trap)
	}
	if result != nil {
		panic("wrong result")
	}
}

func TestFuncWrapRet1(t *testing.T) {
	store := NewStore(NewEngine())
	f := WrapFunc(store, func(c *Caller) int32 {
		return 1
	})
	result, trap := f.Call(store)
	if trap != nil {
		panic(trap)
	}
	if result.(int32) != 1 {
		panic("wrong result")
	}
}

func TestFuncWrapRet2(t *testing.T) {
	store := NewStore(NewEngine())
	f := WrapFunc(store, func(c *Caller) (int64, float64) {
		return 5, 6
	})
	result, trap := f.Call(store)
	if trap != nil {
		panic(trap)
	}
	results := result.([]Val)
	if len(results) != 2 {
		panic("wrong result")
	}
	if results[0].I64() != 5 {
		panic("wrong result")
	}
	if results[1].F64() != 6 {
		panic("wrong result")
	}
}

func TestFuncWrapRetError(t *testing.T) {
	store := NewStore(NewEngine())
	f := WrapFunc(store, func(c *Caller) *Trap {
		return nil
	})
	result, trap := f.Call(store)
	if trap != nil {
		panic(trap)
	}
	if result != nil {
		panic("wrong result")
	}
}

func TestFuncWrapRetErrorTrap(t *testing.T) {
	store := NewStore(NewEngine())
	f := WrapFunc(store, func(c *Caller) *Trap {
		return NewTrap("x")
	})
	_, err := f.Call(store)
	if err == nil {
		panic("expected trap")
	}
	trap := err.(*Trap)
	if trap.Message() != "x" {
		panic("wrong trap")
	}
}

func TestFuncWrapMultiRetWithTrap(t *testing.T) {
	store := NewStore(NewEngine())
	f := WrapFunc(store, func(c *Caller) (int32, float32, *Trap) {
		return 1, 2, nil
	})
	_, err := f.Call(store)
	if err != nil {
		panic(err)
	}
}

func TestFuncWrapPanic(t *testing.T) {
	store := NewStore(NewEngine())
	f := WrapFunc(store, func() { panic("x") })
	var caught interface{}
	var results interface{}
	var err error
	func() {
		defer func() { caught = recover() }()
		results, err = f.Call(store)
	}()
	if caught == nil {
		panic("panic didn't work")
	}
	if caught.(string) != "x" {
		panic("value didn't propagate")
	}
	if err != nil {
		panic(err)
	}
	if results != nil {
		panic("bad results")
	}
}

func TestWrongArgsPanic(t *testing.T) {
	store := NewStore(NewEngine())
	i32 := WrapFunc(store, func(a int32) {})
	i32.Call(store, 1)
	i32.Call(store, int32(1))
	i32.Call(store, ValI32(1))
	_, err := i32.Call(store)
	if err == nil {
		panic("expected error")
	}
	_, err = i32.Call(store, 1, 2)
	if err == nil {
		panic("expected error")
	}
	_, err = i32.Call(store, int64(1))
	if err == nil {
		panic("expected error")
	}
	_, err = i32.Call(store, float32(1))
	if err == nil {
		panic("expected error")
	}
	_, err = i32.Call(store, float64(1))
	if err == nil {
		panic("expected error")
	}
	_, err = i32.Call(store, float32(1))
	if err == nil {
		panic("expected error")
	}
	_, err = i32.Call(store, float64(1))
	if err == nil {
		panic("expected error")
	}
	_, err = i32.Call(store, ValI64(1))
	if err == nil {
		panic("expected error")
	}
	_, err = i32.Call(store, ValF32(1))
	if err == nil {
		panic("expected error")
	}
	_, err = i32.Call(store, ValF64(1))
	if err == nil {
		panic("expected error")
	}

	i64 := WrapFunc(store, func(a int64) {})
	i64.Call(store, 1)
	i64.Call(store, int64(1))
	i64.Call(store, ValI64(1))
	_, err = i64.Call(store, int32(1))
	if err == nil {
		panic("expected error")
	}
	_, err = i64.Call(store, float32(1))
	if err == nil {
		panic("expected error")
	}
	_, err = i64.Call(store, float64(1))
	if err == nil {
		panic("expected error")
	}
	_, err = i64.Call(store, float32(1))
	if err == nil {
		panic("expected error")
	}
	_, err = i64.Call(store, float64(1))
	if err == nil {
		panic("expected error")
	}
	_, err = i64.Call(store, ValI32(1))
	if err == nil {
		panic("expected error")
	}
	_, err = i64.Call(store, ValF32(1))
	if err == nil {
		panic("expected error")
	}
	_, err = i64.Call(store, ValF64(1))
	if err == nil {
		panic("expected error")
	}

	f32 := WrapFunc(store, func(a float32) {})
	f32.Call(store, float32(1))
	_, err = f32.Call(store, 1)
	if err == nil {
		panic("expected error")
	}
	_, err = f32.Call(store, f32)
	if err == nil {
		panic("expected error")
	}
	if len(f32.Type(store).Params()) != 1 {
		panic("wrong param arity")
	}
}

func assertPanic(f func()) {
	var catch interface{}
	func() {
		defer func() { catch = recover() }()
		f()
	}()
	if catch == nil {
		panic("closure didn't panic")
	}

}

func TestNotCallable(t *testing.T) {
	store := NewStore(NewEngine())
	assertPanic(func() { WrapFunc(store, 1) })
}

func TestInterestingTypes(t *testing.T) {
	store := NewStore(NewEngine())
	WrapFunc(store, func(*Store) {})
	WrapFunc(store, func() *Store { return nil })
}

type i32alias int32

const (
	i32_0 i32alias = 0
	i32_1 i32alias = 1
	i32_2 i32alias = 2
)

func TestFuncWrapAliasRet(t *testing.T) {
	store := NewStore(NewEngine())
	f := WrapFunc(store, func(which i32alias) i32alias {
		return which
	})
	result, trap := f.Call(store, i32_1)
	if trap != nil {
		panic(trap)
	}

	if result != i32_1 {
		panic(fmt.Sprintf("wrong result, expected %q, got %q", i32_1, result))
	}
}

func TestCallFuncFromCaller(t *testing.T) {
	wasm, err := Wat2Wasm(`
	(module
		(import "" "f2" (func $f2))
		(func (export "f1")
			(call $f2)
			(call $f2))
		(func (export "f3")
			(nop)))
	`)
	check(err)

	store := NewStore(NewEngine())

	f := NewFunc(store, NewFuncType(nil, nil), func(c *Caller, args []Val) ([]Val, *Trap) {
		fn := c.GetExport("f3").Func()
		_, err := fn.Call(store)
		check(err)
		return nil, nil
	})

	module, err := NewModule(store.Engine, wasm)
	check(err)

	instance, err := NewInstance(store, module, []AsExtern{f})
	check(err)

	_, err = instance.GetExport(store, "f1").Func().Call(store)
	check(err)
}
