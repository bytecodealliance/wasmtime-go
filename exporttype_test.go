package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExportType(t *testing.T) {
	et := NewExportType("x", NewMemoryType(0, false, 0, false))
	require.Equal(t, et.Name(), "x", "bad name")
	require.NotNil(t, et.Type().MemoryType(), "bad type")
}
