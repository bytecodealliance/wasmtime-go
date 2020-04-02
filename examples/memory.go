package main

// An example of how to interact with wasm memory.
//
// Here a small wasm module is used to show how memory is initialized, how to
// read and write memory through the `Memory` object, and how wasm functions
// can trap when dealing with out-of-bounds addresses.

import (
	"github.com/alexcrichton/wasmtime-go/pkg/wasmtime"
	"unsafe"
)

func main() {
	// Create our `Store` context and then compile a module and create an
	// instance from the compiled module all in one go.
	wasmtime_store := wasmtime.NewStore(wasmtime.NewEngine())
	module, err := wasmtime.NewModuleFromFile(wasmtime_store, "examples/memory.wat")
	check(err)
	instance, err := wasmtime.NewInstance(module, []*wasmtime.Extern{})
	check(err)

	// Load up our exports from the instance
	memory := instance.GetExport("memory").Memory()
	size := instance.GetExport("size").Func()
	load := instance.GetExport("load").Func()
	store := instance.GetExport("store").Func()

	println("Checking memory...")
	assert(memory.Size() == 2)
	assert(memory.DataSize() == 0x20000)

	assert(readMemory(memory, 0) == 0)
	assert(readMemory(memory, 0x1000) == 1)
	assert(readMemory(memory, 0x1003) == 4)

	assert(call32(size) == 2)
	assert(call32(load, 0) == 0)
	assert(call32(load, 0x1000) == 1)
	assert(call32(load, 0x1003) == 4)
	assert(call32(load, 0x1ffff) == 0)
	assertTraps(load, 0x20000)

	println("Mutating memory...")
	writeMemory(memory, 0x1003, 5)
	call(store, 0x1002, 6)
	assertTraps(store, 0x20000, 0)

	assert(readMemory(memory, 0x1002) == 6)
	assert(readMemory(memory, 0x1003) == 5)
	assert(call32(load, 0x1002) == 6)
	assert(call32(load, 0x1003) == 5)

	println("Growing memory...")
	assert(memory.Grow(1))
	assert(memory.Size() == 3)
	assert(memory.DataSize() == 0x30000)

	assert(call32(load, 0x20000) == 0)
	call(store, 0x20000, 0)
	assertTraps(load, 0x30000)
	assertTraps(store, 0x30000, 0)

	// Memory can fail to grow
	assert(!memory.Grow(1))
	assert(memory.Grow(0))

	println("Creating stand-alone memory...")
	memorytype := wasmtime.NewMemoryType(wasmtime.Limits{Min: 5, Max: 5})
	memory2 := wasmtime.NewMemory(wasmtime_store, memorytype)
	assert(memory2.Size() == 5)
	assert(!memory2.Grow(1))
	assert(memory2.Grow(0))
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func assert(b bool) {
	if !b {
		panic("assertion failed")
	}
}

func call32(f *wasmtime.Func, args ...interface{}) int32 {
	ret, err := f.Call(args...)
	check(err)
	return ret.(int32)
}

func call(f *wasmtime.Func, args ...interface{}) {
	_, err := f.Call(args...)
	check(err)
}

func assertTraps(f *wasmtime.Func, args ...interface{}) {
	_, err := f.Call(args...)
	_, ok := err.(*wasmtime.Trap)
	if !ok {
		panic("expected a trap")
	}
}

func readMemory(mem *wasmtime.Memory, offset uintptr) uint8 {
	// Note that `Data` returned is a raw `unsafe.Pointer` into wasm memory, care
	// must be taken when using it!
	assert(offset < mem.DataSize())
	return *(*uint8)(unsafe.Pointer(uintptr(mem.Data()) + offset))
}

func writeMemory(mem *wasmtime.Memory, offset uintptr, val uint8) {
	assert(offset < mem.DataSize())
	*(*uint8)(unsafe.Pointer(uintptr(mem.Data()) + offset)) = val
}
