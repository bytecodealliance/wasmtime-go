package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExportType(t *testing.T) {
	et := NewExportType("x", NewMemoryType(0, false, 0))
	assert.Equal(t, et.Name(), "x", "bad name")
	assert.NotNil(t, et.Type().MemoryType(), "bad type")
}
