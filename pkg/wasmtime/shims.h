#include <wasm.h>
#include <wasmtime.h>

wasm_func_t *c_func_new_with_env(wasm_store_t *store, wasm_functype_t *ty, size_t env, int wrap);
wasm_extern_t* go_caller_export_get(const wasmtime_caller_t* caller, char *name_ptr, size_t name_len);
bool go_linker_define(
    wasmtime_linker_t *linker,
    char *module_ptr,
    size_t module_len,
    char *name_ptr,
    size_t name_len,
    wasm_extern_t *item
);
bool go_linker_define_instance(
    wasmtime_linker_t *linker,
    char *name_ptr,
    size_t name_len,
    wasm_instance_t *item
);
