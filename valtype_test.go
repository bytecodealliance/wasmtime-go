package wasmtime

import "testing"

func TestValType(t *testing.T) {
	NewValType(KindI32)
	NewValType(KindI64)
	NewValType(KindF32)
	NewValType(KindF64)
	NewValType(KindExternref)
	NewValType(KindFuncref)
}

func TestValTypeKind(t *testing.T) {
	if NewValType(KindI32).Kind() != KindI32 {
		t.Fatalf("wrong kind")
	}
	if NewValType(KindI64).Kind() != KindI64 {
		t.Fatalf("wrong kind")
	}
	if NewValType(KindF32).Kind() != KindF32 {
		t.Fatalf("wrong kind")
	}
	if NewValType(KindF64).Kind() != KindF64 {
		t.Fatalf("wrong kind")
	}
	if NewValType(KindExternref).Kind() != KindExternref {
		t.Fatalf("wrong kind")
	}
	if NewValType(KindFuncref).Kind() != KindFuncref {
		t.Fatalf("wrong kind")
	}
	if KindI32 == KindI64 {
		t.Fatalf("unequal kinds equal")
	}
	if KindI32 != KindI32 {
		t.Fatalf("equal kinds unequal")
	}
}
