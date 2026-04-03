package wasmtime

// #include <wasmtime.h>
import "C"
import "runtime"

// DefineWasi links a WASI module into this linker, ensuring that all exported functions
// are available for linking.
//
// Returns an error if shadowing is disabled and names are already defined.
func (l *Linker) DefineWasi() error {
	err := C.wasmtime_linker_define_wasi(l.ptr())
	runtime.KeepAlive(l)
	if err == nil {
		return nil
	}

	return mkError(err)
}
