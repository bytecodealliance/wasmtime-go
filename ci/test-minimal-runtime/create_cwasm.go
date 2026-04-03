//go:build ignore

package main

import (
	"log"
	"os"

	wasmtime "github.com/bytecodealliance/wasmtime-go/v43"
)

func main() {
	wasm, err := wasmtime.Wat2Wasm(`(module (func (export "test") (result i32) (i32.const 1)))`)
	check(err)

	cfg := wasmtime.NewConfig()
	cfg.SetGCSupport(false)
	cfg.SetWasmThreads(false)
	cfg.SetWasmComponentModel(false)
	engine := wasmtime.NewEngineWithConfig(cfg)
	module, err := wasmtime.NewModule(engine, wasm)
	check(err)
	defer module.Close()

	artifact, err := module.Serialize()
	check(err)

	check(os.WriteFile("module.cwasm", artifact, 0o644))
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
