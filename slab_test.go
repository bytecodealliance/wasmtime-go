package wasmtime

import (
	"testing"
)

func TestSlab(t *testing.T) {
	var slab slab
	if slab.allocate() != 0 {
		panic("bad alloc")
	}
	if slab.allocate() != 1 {
		panic("bad alloc")
	}
	slab.deallocate(0)
	if slab.allocate() != 0 {
		panic("bad alloc")
	}
	if slab.allocate() != 2 {
		panic("bad alloc")
	}
}
