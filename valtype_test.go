package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValType(t *testing.T) {
	NewValType(KindI32)
	NewValType(KindI64)
	NewValType(KindF32)
	NewValType(KindF64)
	NewValType(KindExternref)
	NewValType(KindFuncref)
}

func TestValTypeKind(t *testing.T) {
	require.Equal(t, KindI32, NewValType(KindI32).Kind(), "wrong kind")
	require.Equal(t, NewValType(KindF32).Kind(), KindF32, "wrong kind")
	require.Equal(t, NewValType(KindF64).Kind(), KindF64, "wrong kind")
	require.Equal(t, NewValType(KindExternref).Kind(), KindExternref, "wrong kind")
	require.Equal(t, NewValType(KindFuncref).Kind(), KindFuncref, "wrong kind")
	require.NotEqual(t, KindI32, KindI64, "unequal kinds equal")
	require.Equal(t, KindI32, KindI32, "equal kinds unequal")
}
