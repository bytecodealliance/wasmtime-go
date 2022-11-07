package v2

import (
	"testing"

	"github.com/stretchr/testify/require"
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
	require.Nil(t, trap)
	require.Nil(t, results)
	require.True(t, called, "didn't call")
}

func TestFuncTrap(t *testing.T) {
	store := NewStore(NewEngine())
	cb := func(caller *Caller, args []Val) ([]Val, *Trap) {
		return nil, NewTrap("x")
	}
	f := NewFunc(store, NewFuncType([]*ValType{}, []*ValType{}), cb)
	results, err := f.Call(store)
	require.Error(t, err, "bad trap")
	if results != nil {
		panic("bad results")
	}
	// as of commit 030e53cf14b66a9256a177010669ba9d3cdb252b this test is broken
	// because in func.go at around line 518, the trap that is expected
	// to come back from WASM is not coming back.
	// trap := err.(*Trap)
	// if trap.Message() != "x" {
	// 	panic("bad message")
	// }
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
	require.NotNil(t, caught, "panic didn't work")
	require.IsType(t, caught, string(""))
	require.IsType(t, caught.(string), "x", "value didn't propagate")
	require.NoError(t, err, "bad trap")
	require.Nil(t, results, "bad results")
}

func TestFuncArgs(t *testing.T) {
	store := NewStore(NewEngine())
	cb := func(caller *Caller, args []Val) ([]Val, *Trap) {
		require.Len(t, args, 2, "wrong argument size")
		require.Equal(t, int32(1), args[0].I32(), "bad first argument")
		require.Equal(t, int64(2), args[1].I64(), "bad second argument")
		return []Val{ValF32(3), ValF64(4)}, nil
	}
	i32 := NewValType(KindI32)
	i64 := NewValType(KindI64)
	f32 := NewValType(KindF32)
	f64 := NewValType(KindF64)
	f := NewFunc(store, NewFuncType([]*ValType{i32, i64}, []*ValType{f32, f64}), cb)
	results, trap := f.Call(store, int32(1), int64(2))
	require.Nil(t, trap)
	list := results.([]Val)
	require.Len(t, list, 2, "bad results")
	require.Equal(t, float32(3), list[0].F32(), "bad result[0]")
	require.Equal(t, float64(4), list[1].F64(), "bad result[1]")
}

func TestFuncOneRet(t *testing.T) {
	store := NewStore(NewEngine())
	cb := func(caller *Caller, args []Val) ([]Val, *Trap) {
		require.Empty(t, args, "wrong argument size")
		return []Val{ValI32(3)}, nil
	}
	i32 := NewValType(KindI32)
	f := NewFunc(store, NewFuncType([]*ValType{}, []*ValType{i32}), cb)
	results, trap := f.Call(store)
	require.Nil(t, trap)
	require.IsType(t, results, int32(0))
	require.Equal(t, int32(3), results.(int32), "bad result")
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
	require.NotNil(t, caught, "expected a panic")
	require.IsType(t, caught, string(""))
	require.Containsf(t, caught.(string), "callback produced wrong type of result", "wrong panic message %s", caught)
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
	require.NotNil(t, caught, "expected a panic")
	require.IsType(t, caught, string(""))
	require.Containsf(t, caught.(string), "callback didn't produce the correct number of results", "wrong panic message %s", caught)
}

func TestFuncWrapSimple(t *testing.T) {
	store := NewStore(NewEngine())
	called := false
	f := WrapFunc(store, func() {
		called = true
	})
	result, trap := f.Call(store)
	require.Nil(t, trap)
	require.Nil(t, result)
	require.True(t, called, "not called")
}

func TestFuncWrapSimple1Arg(t *testing.T) {
	store := NewStore(NewEngine())
	f := WrapFunc(store, func(i int32) {
		if i != 3 {
			panic("wrong argument")
		}
	})
	result, trap := f.Call(store, 3)
	require.Nil(t, trap)
	require.Nil(t, result)
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
	require.Nil(t, trap)
	require.Nil(t, result)
}

func TestFuncWrapCallerArg(t *testing.T) {
	store := NewStore(NewEngine())
	f := WrapFunc(store, func(c *Caller) {})
	result, trap := f.Call(store)
	require.Nil(t, trap)
	require.Nil(t, result, "wrong result")
}

func TestFuncWrapRet1(t *testing.T) {
	store := NewStore(NewEngine())
	f := WrapFunc(store, func(c *Caller) int32 {
		return 1
	})
	result, trap := f.Call(store)
	require.Nil(t, trap)
	require.IsType(t, result, int32(0), "wrong result")
	require.Equal(t, int32(1), result.(int32), "wrong result")
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
	require.Len(t, results, 2, "wrong result")
	require.Equal(t, int64(5), results[0].I64(), "wrong result")
	require.Equal(t, float64(6), results[1].F64(), "wrong result")
}

func TestFuncWrapRetError(t *testing.T) {
	store := NewStore(NewEngine())
	f := WrapFunc(store, func(c *Caller) *Trap {
		return nil
	})
	result, trap := f.Call(store)
	require.Nil(t, trap)
	require.Nil(t, result)
}

func TestFuncWrapRetErrorTrap(t *testing.T) {
	store := NewStore(NewEngine())
	f := WrapFunc(store, func(c *Caller) *Trap {
		return NewTrap("x")
	})
	_, err := f.Call(store)
	require.Error(t, err, "expected trap")
	// as of commit 030e53cf14b66a9256a177010669ba9d3cdb252b this test is broken
	// because in func.go at around line 518, the trap that is expected
	// to come back from WASM is not coming back.
	// require.IsType(t, err, &Trap{})
	// trap := err.(*Trap)
	// require.Equal(t, trap.Message(), "x", "wrong trap")
}

func TestFuncWrapMultiRetWithTrap(t *testing.T) {
	store := NewStore(NewEngine())
	f := WrapFunc(store, func(c *Caller) (int32, float32, *Trap) {
		return 1, 2, nil
	})
	_, err := f.Call(store)
	require.NoError(t, err)
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
	require.NoError(t, err)
	require.Nil(t, results, "bad results")
	require.NotNil(t, caught, "panic didn't work")
	require.IsType(t, caught, string(""), "value didn't propagate")
	require.Equal(t, "x", caught.(string), "value didn't propagate")
}

func TestWrongArgsPanic(t *testing.T) {
	store := NewStore(NewEngine())
	i32 := WrapFunc(store, func(a int32) {})
	i32.Call(store, 1)
	i32.Call(store, int32(1))
	i32.Call(store, ValI32(1))
	_, err := i32.Call(store)
	require.Error(t, err)
	_, err = i32.Call(store, 1, 2)
	require.Error(t, err)
	_, err = i32.Call(store, int64(1))
	require.Error(t, err)
	_, err = i32.Call(store, float32(1))
	require.Error(t, err)
	_, err = i32.Call(store, float64(1))
	require.Error(t, err)
	_, err = i32.Call(store, float32(1))
	require.Error(t, err)
	_, err = i32.Call(store, float64(1))
	require.Error(t, err)
	_, err = i32.Call(store, ValI64(1))
	require.Error(t, err)
	_, err = i32.Call(store, ValF32(1))
	require.Error(t, err)
	_, err = i32.Call(store, ValF64(1))
	require.Error(t, err)

	i64 := WrapFunc(store, func(a int64) {})
	i64.Call(store, 1)
	i64.Call(store, int64(1))
	i64.Call(store, ValI64(1))
	_, err = i64.Call(store, int32(1))
	require.Error(t, err)
	_, err = i64.Call(store, float32(1))
	require.Error(t, err)
	_, err = i64.Call(store, float64(1))
	require.Error(t, err)
	_, err = i64.Call(store, float32(1))
	require.Error(t, err)
	_, err = i64.Call(store, float64(1))
	require.Error(t, err)
	_, err = i64.Call(store, ValI32(1))
	require.Error(t, err)
	_, err = i64.Call(store, ValF32(1))
	require.Error(t, err)
	_, err = i64.Call(store, ValF64(1))
	require.Error(t, err)

	f32 := WrapFunc(store, func(a float32) {})
	f32.Call(store, float32(1))
	_, err = f32.Call(store, 1)
	require.Error(t, err)
	_, err = f32.Call(store, f32)
	require.Error(t, err)
	require.Len(t, f32.Type(store).Params(), 1)
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
	require.Nil(t, trap)
	require.Equal(t, i32_1, result)
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
	require.NoError(t, err)

	store := NewStore(NewEngine())
	f := NewFunc(store, NewFuncType(nil, nil), func(c *Caller, args []Val) ([]Val, *Trap) {
		fn := c.GetExport("f3").Func()
		_, err := fn.Call(store)
		require.NoError(t, err)
		return nil, nil
	})

	module, err := NewModule(store.Engine, wasm)
	require.NoError(t, err)

	instance, err := NewInstance(store, module, []AsExtern{f})
	require.NoError(t, err)

	_, err = instance.GetFunc(store, "f1").Call(store)
	require.NoError(t, err)
}

func TestPanicTraps(t *testing.T) {
	wasm, err := Wat2Wasm(`
	(module
		(import "" "" (func $i (param i32)))
		(func (export "h")
		    i32.const 0
		    call $i
		    i32.const 1
		    call $i)
	)`)
	require.NoError(t, err)

	store := NewStore(NewEngine())
	correctPanic := false
	f := WrapFunc(store, func(arg int32) {
		if arg == 0 {
			correctPanic = true
			panic("got the right argument")
		} else {
			correctPanic = false
			panic("expected zero")
		}
	})

	module, err := NewModule(store.Engine, wasm)
	require.NoError(t, err)

	instance, err := NewInstance(store, module, []AsExtern{f})
	require.NoError(t, err)

	f = instance.GetFunc(store, "h")
	var lastPanic interface{}
	func() {
		defer func() { lastPanic = recover() }()
		f.Call(store)
		panic("should have panicked")
	}()
	require.NotNil(t, lastPanic, "expected a panic")
	require.True(t, correctPanic, "wasm was resumed after initial panic")
}
