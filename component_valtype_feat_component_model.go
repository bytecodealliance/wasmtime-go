package wasmtime

// #include <wasmtime.h>
//
// static inline uint8_t go_component_valtype_kind(const wasmtime_component_valtype_t *vt) {
//   return vt->kind;
// }
import "C"

import "runtime"

// ComponentValTypeKind discriminates the WIT type that a [ComponentValType]
// represents. Only constants for the 13 primitive kinds (bool through
// string) are exposed in this release; constants for the composite kinds
// (list / record / tuple / variant / enum / option / result / flags / own /
// borrow / future / stream / error-context / map) are intentionally
// commented out below — they will be uncommented when each composite kind
// gets a dedicated payload accessor and a test path that exercises it.
type ComponentValTypeKind uint8

const (
	ComponentValTypeKindBool   ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_BOOL
	ComponentValTypeKindS8     ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_S8
	ComponentValTypeKindS16    ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_S16
	ComponentValTypeKindS32    ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_S32
	ComponentValTypeKindS64    ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_S64
	ComponentValTypeKindU8     ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_U8
	ComponentValTypeKindU16    ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_U16
	ComponentValTypeKindU32    ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_U32
	ComponentValTypeKindU64    ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_U64
	ComponentValTypeKindF32    ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_F32
	ComponentValTypeKindF64    ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_F64
	ComponentValTypeKindChar   ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_CHAR
	ComponentValTypeKindString ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_STRING

	// Composite-kind constants are deferred until each gets a payload
	// accessor that returns the corresponding sub-type wrapper, with a
	// test path. Uncomment as each one lands.
	//
	// ComponentValTypeKindList         ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_LIST
	// ComponentValTypeKindRecord       ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_RECORD
	// ComponentValTypeKindTuple        ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_TUPLE
	// ComponentValTypeKindVariant      ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_VARIANT
	// ComponentValTypeKindEnum         ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_ENUM
	// ComponentValTypeKindOption       ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_OPTION
	// ComponentValTypeKindResult       ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_RESULT
	// ComponentValTypeKindFlags        ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_FLAGS
	// ComponentValTypeKindOwn          ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_OWN
	// ComponentValTypeKindBorrow       ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_BORROW
	// ComponentValTypeKindFuture       ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_FUTURE
	// ComponentValTypeKindStream       ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_STREAM
	// ComponentValTypeKindErrorContext ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_ERROR_CONTEXT
	// ComponentValTypeKindMap          ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_MAP
)

// ComponentValType describes the WIT type of a value in the component
// model.
//
// In this release [ComponentValType.Kind] returns one of the 13 primitive
// [ComponentValTypeKind] constants. If the underlying value type is a
// composite kind, the returned uint8 will not match any exposed constant
// — those, along with payload accessors that return their sub-type
// wrappers, arrive in follow-up work.
type ComponentValType struct {
	val    C.wasmtime_component_valtype_t
	closed bool
}

func mkComponentValType(val C.wasmtime_component_valtype_t) *ComponentValType {
	vt := &ComponentValType{val: val}
	runtime.SetFinalizer(vt, func(vt *ComponentValType) {
		vt.Close()
	})
	return vt
}

// Kind returns the discriminator of this value type.
func (vt *ComponentValType) Kind() ComponentValTypeKind {
	if vt.closed {
		panic("object has been closed already")
	}
	maybeGC()
	return ComponentValTypeKind(C.go_component_valtype_kind(&vt.val))
}

// Close deallocates this value type explicitly.
func (vt *ComponentValType) Close() {
	if vt.closed {
		return
	}
	runtime.SetFinalizer(vt, nil)
	C.wasmtime_component_valtype_delete(&vt.val)
	vt.closed = true
}
