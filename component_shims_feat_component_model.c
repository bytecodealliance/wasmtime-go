//go:build wasmtime_component_model

#include "_cgo_export.h"
#include "component_shims_feat_component_model.h"

static wasmtime_error_t *component_func_callback(
    void *env,
    wasmtime_context_t *context,
    const wasmtime_component_func_type_t *ty,
    wasmtime_component_val_t *args,
    size_t nargs,
    wasmtime_component_val_t *results,
    size_t nresults) {
  return goComponentFuncCallback(
      env, context, (wasmtime_component_func_type_t *)ty,
      args, nargs, results, nresults);
}

wasmtime_error_t *go_component_linker_instance_add_func(
    wasmtime_component_linker_instance_t *instance,
    const char *name,
    size_t name_len,
    size_t env) {
  return wasmtime_component_linker_instance_add_func(
      instance, name, name_len, component_func_callback, (void *)env,
      goFinalizeComponentFunc);
}

static wasmtime_error_t *component_resource_destructor(
    void *env, wasmtime_context_t *context, uint32_t rep) {
  return goComponentResourceDestructor(env, context, rep);
}

wasmtime_error_t *go_component_linker_instance_add_resource(
    wasmtime_component_linker_instance_t *instance,
    const char *name,
    size_t name_len,
    const wasmtime_component_resource_type_t *resource,
    size_t env) {
  return wasmtime_component_linker_instance_add_resource(
      instance, name, name_len, resource, component_resource_destructor,
      (void *)env, goFinalizeComponentResourceDestructor);
}
