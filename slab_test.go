package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSlab(t *testing.T) {
	var slab slab
	require.Equal(t, 0, slab.allocate())
	require.Equal(t, 1, slab.allocate())
	slab.deallocate(0)
	require.Equal(t, 0, slab.allocate())
	require.Equal(t, 2, slab.allocate())
}
