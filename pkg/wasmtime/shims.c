#include "_cgo_export.h"
#include "shims.h"

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

wasm_extern_t* go_caller_export_get(
  const wasmtime_caller_t* caller,
  char *name_ptr,
  size_t name_len
) {
  wasm_byte_vec_t name;
  name.data = name_ptr;
  name.size = name_len;
  return wasmtime_caller_export_get(caller, &name);
}

bool go_linker_define(
    wasmtime_linker_t *linker,
    char *module_ptr,
    size_t module_len,
    char *name_ptr,
    size_t name_len,
    wasm_extern_t *item
) {
  wasm_byte_vec_t module;
  module.data = module_ptr;
  module.size = module_len;
  wasm_byte_vec_t name;
  name.data = name_ptr;
  name.size = name_len;
  return wasmtime_linker_define(linker, &module, &name, item);
}

bool go_linker_define_instance(
    wasmtime_linker_t *linker,
    char *name_ptr,
    size_t name_len,
    wasm_instance_t *instance
) {
  wasm_byte_vec_t name;
  name.data = name_ptr;
  name.size = name_len;
  return wasmtime_linker_define_instance(linker, &name, instance);
}
