package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlab(t *testing.T) {
	var slab slab
	assert.Equal(t, 0, slab.allocate())
	assert.Equal(t, 1, slab.allocate())
	slab.deallocate(0)
	assert.Equal(t, 0, slab.allocate())
	assert.Equal(t, 2, slab.allocate())
}
