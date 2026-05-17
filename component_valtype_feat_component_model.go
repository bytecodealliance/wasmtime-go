package wasmtime

// #include <wasmtime.h>
//
// static inline uint8_t go_component_valtype_kind(const wasmtime_component_valtype_t *vt) {
//   return vt->kind;
// }
// static inline wasmtime_component_list_type_t *go_component_valtype_list(const wasmtime_component_valtype_t *vt) {
//   return vt->of.list;
// }
// static inline wasmtime_component_record_type_t *go_component_valtype_record(const wasmtime_component_valtype_t *vt) {
//   return vt->of.record;
// }
// static inline wasmtime_component_tuple_type_t *go_component_valtype_tuple(const wasmtime_component_valtype_t *vt) {
//   return vt->of.tuple;
// }
// static inline wasmtime_component_enum_type_t *go_component_valtype_enum(const wasmtime_component_valtype_t *vt) {
//   return vt->of.enum_;
// }
// static inline wasmtime_component_flags_type_t *go_component_valtype_flags(const wasmtime_component_valtype_t *vt) {
//   return vt->of.flags;
// }
import "C"

import "runtime"

// ComponentValTypeKind discriminates the WIT type that a [ComponentValType]
// represents. This release exposes the 13 primitive kinds plus the
// product/enum composite kinds (list / record / tuple / enum / flags).
// Constants for the remaining composite kinds (variant / option / result /
// own / borrow / future / stream / error-context / map) are commented out
// and arrive in follow-up work along with their payload accessors.
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
	ComponentValTypeKindList   ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_LIST
	ComponentValTypeKindRecord ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_RECORD
	ComponentValTypeKindTuple  ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_TUPLE
	ComponentValTypeKindEnum   ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_ENUM
	ComponentValTypeKindFlags  ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_FLAGS

	// Remaining composite-kind constants are deferred until each gets a
	// payload accessor that returns the corresponding sub-type wrapper,
	// with a test path. Uncomment as each one lands.
	//
	// ComponentValTypeKindVariant      ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_VARIANT
	// ComponentValTypeKindOption       ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_OPTION
	// ComponentValTypeKindResult       ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_RESULT
	// ComponentValTypeKindOwn          ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_OWN
	// ComponentValTypeKindBorrow       ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_BORROW
	// ComponentValTypeKindFuture       ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_FUTURE
	// ComponentValTypeKindStream       ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_STREAM
	// ComponentValTypeKindErrorContext ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_ERROR_CONTEXT
	// ComponentValTypeKindMap          ComponentValTypeKind = C.WASMTIME_COMPONENT_VALTYPE_MAP
)

// ComponentValType describes the WIT type of a value in the component
// model. Use [ComponentValType.Kind] to discriminate, then call the
// corresponding downcast method ([ComponentValType.List],
// [ComponentValType.Record], ...) for composite kinds.
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

// List returns the [ComponentListType] wrapper when this value type's kind
// is [ComponentValTypeKindList], or nil otherwise. The returned wrapper has
// an independent lifecycle from the parent.
func (vt *ComponentValType) List() *ComponentListType {
	if vt.Kind() != ComponentValTypeKindList {
		return nil
	}
	cloned := C.wasmtime_component_list_type_clone(C.go_component_valtype_list(&vt.val))
	runtime.KeepAlive(vt)
	return mkComponentListType(cloned)
}

// Record returns the [ComponentRecordType] wrapper when this value type's
// kind is [ComponentValTypeKindRecord], or nil otherwise.
func (vt *ComponentValType) Record() *ComponentRecordType {
	if vt.Kind() != ComponentValTypeKindRecord {
		return nil
	}
	cloned := C.wasmtime_component_record_type_clone(C.go_component_valtype_record(&vt.val))
	runtime.KeepAlive(vt)
	return mkComponentRecordType(cloned)
}

// Tuple returns the [ComponentTupleType] wrapper when this value type's
// kind is [ComponentValTypeKindTuple], or nil otherwise.
func (vt *ComponentValType) Tuple() *ComponentTupleType {
	if vt.Kind() != ComponentValTypeKindTuple {
		return nil
	}
	cloned := C.wasmtime_component_tuple_type_clone(C.go_component_valtype_tuple(&vt.val))
	runtime.KeepAlive(vt)
	return mkComponentTupleType(cloned)
}

// Enum returns the [ComponentEnumType] wrapper when this value type's kind
// is [ComponentValTypeKindEnum], or nil otherwise.
func (vt *ComponentValType) Enum() *ComponentEnumType {
	if vt.Kind() != ComponentValTypeKindEnum {
		return nil
	}
	cloned := C.wasmtime_component_enum_type_clone(C.go_component_valtype_enum(&vt.val))
	runtime.KeepAlive(vt)
	return mkComponentEnumType(cloned)
}

// Flags returns the [ComponentFlagsType] wrapper when this value type's
// kind is [ComponentValTypeKindFlags], or nil otherwise.
func (vt *ComponentValType) Flags() *ComponentFlagsType {
	if vt.Kind() != ComponentValTypeKindFlags {
		return nil
	}
	cloned := C.wasmtime_component_flags_type_clone(C.go_component_valtype_flags(&vt.val))
	runtime.KeepAlive(vt)
	return mkComponentFlagsType(cloned)
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

// ComponentListType is the payload of a `list<T>` value type.
type ComponentListType struct {
	_ptr *C.wasmtime_component_list_type_t
}

func mkComponentListType(p *C.wasmtime_component_list_type_t) *ComponentListType {
	lt := &ComponentListType{_ptr: p}
	runtime.SetFinalizer(lt, func(lt *ComponentListType) {
		lt.Close()
	})
	return lt
}

func (lt *ComponentListType) ptr() *C.wasmtime_component_list_type_t {
	if lt._ptr == nil {
		panic("object has been closed already")
	}
	maybeGC()
	return lt._ptr
}

// Element returns the element type of this list.
func (lt *ComponentListType) Element() *ComponentValType {
	var out C.wasmtime_component_valtype_t
	C.wasmtime_component_list_type_element(lt.ptr(), &out)
	runtime.KeepAlive(lt)
	return mkComponentValType(out)
}

// Close deallocates this list type explicitly.
func (lt *ComponentListType) Close() {
	if lt._ptr == nil {
		return
	}
	runtime.SetFinalizer(lt, nil)
	C.wasmtime_component_list_type_delete(lt._ptr)
	lt._ptr = nil
}

// ComponentRecordType is the payload of a `record { ... }` value type.
type ComponentRecordType struct {
	_ptr *C.wasmtime_component_record_type_t
}

func mkComponentRecordType(p *C.wasmtime_component_record_type_t) *ComponentRecordType {
	rt := &ComponentRecordType{_ptr: p}
	runtime.SetFinalizer(rt, func(rt *ComponentRecordType) {
		rt.Close()
	})
	return rt
}

func (rt *ComponentRecordType) ptr() *C.wasmtime_component_record_type_t {
	if rt._ptr == nil {
		panic("object has been closed already")
	}
	maybeGC()
	return rt._ptr
}

// FieldCount returns the number of fields in this record.
func (rt *ComponentRecordType) FieldCount() int {
	n := C.wasmtime_component_record_type_field_count(rt.ptr())
	runtime.KeepAlive(rt)
	return int(n)
}

// FieldNth returns the name and type of the `i`-th field, or `("", nil)`
// if `i` is out of range.
func (rt *ComponentRecordType) FieldNth(i int) (string, *ComponentValType) {
	var nameP *C.char
	var nameLen C.size_t
	var out C.wasmtime_component_valtype_t
	found := C.wasmtime_component_record_type_field_nth(rt.ptr(), C.size_t(i), &nameP, &nameLen, &out)
	runtime.KeepAlive(rt)
	if !bool(found) {
		return "", nil
	}
	return C.GoStringN(nameP, C.int(nameLen)), mkComponentValType(out)
}

// Close deallocates this record type explicitly.
func (rt *ComponentRecordType) Close() {
	if rt._ptr == nil {
		return
	}
	runtime.SetFinalizer(rt, nil)
	C.wasmtime_component_record_type_delete(rt._ptr)
	rt._ptr = nil
}

// ComponentTupleType is the payload of a `tuple<T1, T2, ...>` value type.
type ComponentTupleType struct {
	_ptr *C.wasmtime_component_tuple_type_t
}

func mkComponentTupleType(p *C.wasmtime_component_tuple_type_t) *ComponentTupleType {
	tt := &ComponentTupleType{_ptr: p}
	runtime.SetFinalizer(tt, func(tt *ComponentTupleType) {
		tt.Close()
	})
	return tt
}

func (tt *ComponentTupleType) ptr() *C.wasmtime_component_tuple_type_t {
	if tt._ptr == nil {
		panic("object has been closed already")
	}
	maybeGC()
	return tt._ptr
}

// TypesCount returns the number of types in this tuple.
func (tt *ComponentTupleType) TypesCount() int {
	n := C.wasmtime_component_tuple_type_types_count(tt.ptr())
	runtime.KeepAlive(tt)
	return int(n)
}

// TypesNth returns the type at position `i`, or nil if `i` is out of range.
func (tt *ComponentTupleType) TypesNth(i int) *ComponentValType {
	var out C.wasmtime_component_valtype_t
	found := C.wasmtime_component_tuple_type_types_nth(tt.ptr(), C.size_t(i), &out)
	runtime.KeepAlive(tt)
	if !bool(found) {
		return nil
	}
	return mkComponentValType(out)
}

// Close deallocates this tuple type explicitly.
func (tt *ComponentTupleType) Close() {
	if tt._ptr == nil {
		return
	}
	runtime.SetFinalizer(tt, nil)
	C.wasmtime_component_tuple_type_delete(tt._ptr)
	tt._ptr = nil
}

// ComponentEnumType is the payload of an `enum "a" "b" ...` value type.
type ComponentEnumType struct {
	_ptr *C.wasmtime_component_enum_type_t
}

func mkComponentEnumType(p *C.wasmtime_component_enum_type_t) *ComponentEnumType {
	et := &ComponentEnumType{_ptr: p}
	runtime.SetFinalizer(et, func(et *ComponentEnumType) {
		et.Close()
	})
	return et
}

func (et *ComponentEnumType) ptr() *C.wasmtime_component_enum_type_t {
	if et._ptr == nil {
		panic("object has been closed already")
	}
	maybeGC()
	return et._ptr
}

// NamesCount returns the number of cases in this enum.
func (et *ComponentEnumType) NamesCount() int {
	n := C.wasmtime_component_enum_type_names_count(et.ptr())
	runtime.KeepAlive(et)
	return int(n)
}

// NamesNth returns the name of the `i`-th case, or "" if `i` is out of range.
func (et *ComponentEnumType) NamesNth(i int) string {
	var nameP *C.char
	var nameLen C.size_t
	found := C.wasmtime_component_enum_type_names_nth(et.ptr(), C.size_t(i), &nameP, &nameLen)
	runtime.KeepAlive(et)
	if !bool(found) {
		return ""
	}
	return C.GoStringN(nameP, C.int(nameLen))
}

// Close deallocates this enum type explicitly.
func (et *ComponentEnumType) Close() {
	if et._ptr == nil {
		return
	}
	runtime.SetFinalizer(et, nil)
	C.wasmtime_component_enum_type_delete(et._ptr)
	et._ptr = nil
}

// ComponentFlagsType is the payload of a `flags "a" "b" ...` value type.
type ComponentFlagsType struct {
	_ptr *C.wasmtime_component_flags_type_t
}

func mkComponentFlagsType(p *C.wasmtime_component_flags_type_t) *ComponentFlagsType {
	ft := &ComponentFlagsType{_ptr: p}
	runtime.SetFinalizer(ft, func(ft *ComponentFlagsType) {
		ft.Close()
	})
	return ft
}

func (ft *ComponentFlagsType) ptr() *C.wasmtime_component_flags_type_t {
	if ft._ptr == nil {
		panic("object has been closed already")
	}
	maybeGC()
	return ft._ptr
}

// NamesCount returns the number of flag names in this flags type.
func (ft *ComponentFlagsType) NamesCount() int {
	n := C.wasmtime_component_flags_type_names_count(ft.ptr())
	runtime.KeepAlive(ft)
	return int(n)
}

// NamesNth returns the name of the `i`-th flag, or "" if `i` is out of range.
func (ft *ComponentFlagsType) NamesNth(i int) string {
	var nameP *C.char
	var nameLen C.size_t
	found := C.wasmtime_component_flags_type_names_nth(ft.ptr(), C.size_t(i), &nameP, &nameLen)
	runtime.KeepAlive(ft)
	if !bool(found) {
		return ""
	}
	return C.GoStringN(nameP, C.int(nameLen))
}

// Close deallocates this flags type explicitly.
func (ft *ComponentFlagsType) Close() {
	if ft._ptr == nil {
		return
	}
	runtime.SetFinalizer(ft, nil)
	C.wasmtime_component_flags_type_delete(ft._ptr)
	ft._ptr = nil
}
