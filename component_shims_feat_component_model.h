//go:build wasmtime_component_model

#include <wasmtime.h>

wasmtime_error_t *go_component_linker_instance_add_func(
    wasmtime_component_linker_instance_t *instance,
    const char *name,
    size_t name_len,
    size_t env);

wasmtime_error_t *go_component_linker_instance_add_resource(
    wasmtime_component_linker_instance_t *instance,
    const char *name,
    size_t name_len,
    const wasmtime_component_resource_type_t *resource,
    size_t env);
