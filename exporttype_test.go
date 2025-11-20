package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExportType(t *testing.T) {
	mt, err := NewMemoryType(0, false, 0, false)
	require.NoError(t, err)
	et := NewExportType("x", mt)
	require.Equal(t, et.Name(), "x", "bad name")
	require.NotNil(t, et.Type().MemoryType(), "bad type")
}
