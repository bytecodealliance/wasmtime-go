#include "_cgo_export.h"
#include "func.h"
#include <wasmtime.h>

static wasm_trap_t* trampoline(
   const wasmtime_caller_t *caller,
   void *env,
   const wasm_val_t *args,
   wasm_val_t *results
) {
    return goTrampolineNew((wasmtime_caller_t*) caller, (size_t) env, (wasm_val_t*) args, results);
}

static wasm_trap_t* wrap_trampoline(
   const wasmtime_caller_t *caller,
   void *env,
   const wasm_val_t *args,
   wasm_val_t *results
) {
    return goTrampolineWrap((wasmtime_caller_t*) caller, (size_t) env, (wasm_val_t*) args, results);
}

wasm_func_t *c_func_new_with_env(wasm_store_t *store, wasm_functype_t *ty, size_t env, int wrap) {
  if (wrap)
    return wasmtime_func_new_with_env(store, ty, wrap_trampoline, (void*) env, goFinalizeWrap);
  return wasmtime_func_new_with_env(store, ty, trampoline, (void*) env, goFinalizeNew);
}
