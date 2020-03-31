#include "_cgo_export.h"
#include "func.h"
#include <wasmtime.h>

static wasm_trap_t* trampoline(
   const wasmtime_caller_t *caller,
   void *env,
   const wasm_val_t *args,
   wasm_val_t *results
) {
    return goTrampoline((wasmtime_caller_t*) caller, env, (wasm_val_t*) args, results);
}

wasm_func_t *c_func_new_with_env(wasm_store_t *store, wasm_functype_t *ty, void *env) {
  return wasmtime_func_new_with_env(store, ty, trampoline, env, goFinalize);
}
