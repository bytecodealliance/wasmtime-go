package wasmtime

import "testing"
import "github.com/stretchr/testify/require"

func TestWasiConfig(t *testing.T) {
	config := NewWasiConfig()
	defer config.Close()
	config.SetEnv([]string{"WASMTIME"}, []string{"GO"})
	err := config.PreopenDir(".", ".", DIR_READ, FILE_READ)
	require.Nil(t, err)
	err = config.PreopenDir(".", ".", DIR_READ|DIR_WRITE, FILE_READ|FILE_WRITE)
	require.Nil(t, err)

}
