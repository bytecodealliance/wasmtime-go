load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")
load("@bazel_gazelle//:def.bzl", "gazelle")

# gazelle:prefix github.com/bytecodealliance/wasmtime-go
gazelle(name = "gazelle")

cc_import(
    name = "wasmtime",
    static_library = select({
        "@io_bazel_rules_go//go/platform:darwin": "build/macos-x86_64/libwasmtime.a",
        "@io_bazel_rules_go//go/platform:linux_amd64": "build/linux-x86_64/libwasmtime.a",
        "@io_bazel_rules_go//go/platform:windows_amd64": "build/windows-x86_64/libwasmtime.a",
    }),
    hdrs = glob(["build/include/*.h"]),
    visibility = ["//visibility:public"]
)

go_library(
    name = "go_default_library",
    srcs = [
        "config.go",
        "doc.go",
        "engine.go",
        "error.go",
        "exporttype.go",
        "extern.go",
        "externtype.go",
        "ffi.go",
        "freelist.go",
        "func.go",
        "functype.go",
        "global.go",
        "globaltype.go",
        "importtype.go",
        "instance.go",
        "limits.go",
        "linker.go",
        "maybe_gc_no.go",
        "memory.go",
        "memorytype.go",
        "module.go",
        "shims.c",
        "shims.h",
        "slab.go",
        "store.go",
        "table.go",
        "tabletype.go",
        "trap.go",
        "val.go",
        "valtype.go",
        "wasi.go",
        "wat2wasm.go",
    ],
    cgo = True,
    clinkopts = select({
        "@io_bazel_rules_go//go/platform:aix": [
            "-lwasmtime -lm -ldl",
        ],
        "@io_bazel_rules_go//go/platform:android": [
            "-lwasmtime -lm -ldl",
        ],
        "@io_bazel_rules_go//go/platform:darwin": [
            "-lwasmtime -lm -ldl",
        ],
        "@io_bazel_rules_go//go/platform:dragonfly": [
            "-lwasmtime -lm -ldl",
        ],
        "@io_bazel_rules_go//go/platform:freebsd": [
            "-lwasmtime -lm -ldl",
        ],
        "@io_bazel_rules_go//go/platform:illumos": [
            "-lwasmtime -lm -ldl",
        ],
        "@io_bazel_rules_go//go/platform:ios": [
            "-lwasmtime -lm -ldl",
        ],
        "@io_bazel_rules_go//go/platform:js": [
            "-lwasmtime -lm -ldl",
        ],
        "@io_bazel_rules_go//go/platform:linux": [
            "-lwasmtime -lm -ldl",
        ],
        "@io_bazel_rules_go//go/platform:nacl": [
            "-lwasmtime -lm -ldl",
        ],
        "@io_bazel_rules_go//go/platform:netbsd": [
            "-lwasmtime -lm -ldl",
        ],
        "@io_bazel_rules_go//go/platform:openbsd": [
            "-lwasmtime -lm -ldl",
        ],
        "@io_bazel_rules_go//go/platform:plan9": [
            "-lwasmtime -lm -ldl",
        ],
        "@io_bazel_rules_go//go/platform:solaris": [
            "-lwasmtime -lm -ldl",
        ],
        "@io_bazel_rules_go//go/platform:windows": [
            "-lwasmtime -luserenv -lole32 -lntdll -lws2_32 -lkernel32",
        ],
        "//conditions:default": [],
    }) + select({
        "@io_bazel_rules_go//go/platform:android_amd64": [
            "-Lbuild/linux-x86_64",
        ],
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "-Lbuild/macos-x86_64",
        ],
        "@io_bazel_rules_go//go/platform:ios_amd64": [
            "-Lbuild/macos-x86_64",
        ],
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            "-Lbuild/linux-x86_64",
        ],
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "-Lbuild/windows-x86_64",
        ],
        "//conditions:default": [],
    }),
    cdeps = [":wasmtime"], # add wasmtime dep
    copts = [
        "-Ibuild/include",
    ] + select({
        "@io_bazel_rules_go//go/platform:windows": [
            "-DWASM_API_EXTERN= -DWASI_API_EXTERN=",
        ],
        "//conditions:default": [],
    }),
    importpath = "github.com/bytecodealliance/wasmtime-go",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "config_test.go",
        "doc_test.go",
        "engine_test.go",
        "exporttype_test.go",
        "func_test.go",
        "functype_test.go",
        "global_test.go",
        "globaltype_test.go",
        "importtype_test.go",
        "instance_test.go",
        "linker_test.go",
        "memorytype_test.go",
        "module_test.go",
        "slab_test.go",
        "store_test.go",
        "table_test.go",
        "tabletype_test.go",
        "trap_test.go",
        "valtype_test.go",
        "wasi_test.go",
        "wasm2wat_test.go",
    ],
    embed = [":go_default_library"],
)
