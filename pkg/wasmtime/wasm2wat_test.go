package wasmtime

import "testing"

func TestWat2Wasm(t *testing.T) {
	wasm, err := Wat2Wasm("(module)")
	if err != nil {
		panic(err)
	}
	if len(wasm) != 8 {
		panic("wrong wasm")
	}
	_, err = Wat2Wasm("___")
	if err == nil {
		panic("expected an error")
	}
}
