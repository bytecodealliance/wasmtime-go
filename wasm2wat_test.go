package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWat2Wasm(t *testing.T) {
	wasm, err := Wat2Wasm("(module)")
	assert.NoError(t, err)
	assert.Len(t, wasm, 8, "wrong wasm")
	_, err = Wat2Wasm("___")
	assert.Error(t, err, "expected an error")
}
