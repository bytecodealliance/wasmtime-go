package v2

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWat2Wasm(t *testing.T) {
	wasm, err := Wat2Wasm("(module)")
	require.NoError(t, err)
	require.Len(t, wasm, 8, "wrong wasm")
	_, err = Wat2Wasm("___")
	require.Error(t, err, "expected an error")
}
