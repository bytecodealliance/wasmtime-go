package wasmtime

import "testing"
import "fmt"
import "strings"

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
	results, trap := f.Call()
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
		return nil, NewTrap(store, "x")
	}
	f := NewFunc(store, NewFuncType([]*ValType{}, []*ValType{}), cb)
	results, trap := f.Call()
	if trap == nil {
		panic("bad trap")
	}
	if results != nil {
		panic("bad results")
	}
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
	var trap *Trap
	func() {
		defer func() { caught = recover() }()
		results, trap = f.Call()
	}()
	if caught == nil {
		panic("panic didn't work")
	}
	if caught.(string) != "x" {
		panic("value didn't propagate")
	}
	if trap != nil {
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
	results, trap := f.Call(int32(1), int64(2))
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
	results, trap := f.Call()
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
		f.Call()
	}()
	if caught == nil {
		panic("expected a panic")
	}
	if !strings.Contains(caught.(string), "callback produced wrong type of result") {
		panic(fmt.Sprintf("wrong panic message %s", caught))
	}
}
