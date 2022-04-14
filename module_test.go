package wasmtime

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModule(t *testing.T) {
	_, err := NewModule(NewEngine(), []byte{})
	assert.Error(t, err)
	_, err = NewModule(NewEngine(), []byte{1})
	assert.Error(t, err)
}

func TestModuleValidate(t *testing.T) {
	assert.NotNil(t, ModuleValidate(NewEngine(), []byte{}), "expected an error")
	assert.NotNil(t, ModuleValidate(NewEngine(), []byte{1}), "expected an error")
	wasm, err := Wat2Wasm(`(module)`)
	assert.NoError(t, err)
	assert.Nil(t, ModuleValidate(NewEngine(), wasm), "expected valid module")
}

func TestModuleImports(t *testing.T) {
	wasm, err := Wat2Wasm(`
          (module
            (import "" "f" (func))
            (import "a" "g" (global i32))
            (import "" "" (table 1 funcref))
            (import "" "" (memory 1))
          )
        `)
	assert.NoError(t, err)
	module, err := NewModule(NewEngine(), wasm)
	assert.NoError(t, err)
	imports := module.Imports()
	assert.Len(t, imports, 4)
	assert.Equal(t, "", imports[0].Module())
	assert.Equal(t, "f", *imports[0].Name())
	assert.NotNil(t, imports[0].Type().FuncType())
	assert.Len(t, imports[0].Type().FuncType().Params(), 0)
	assert.Len(t, imports[0].Type().FuncType().Results(), 0)

	assert.Equal(t, "a", imports[1].Module())
	assert.Equal(t, "g", *imports[1].Name())
	assert.NotNil(t, imports[1].Type().GlobalType())
	assert.Equal(t, KindI32, imports[1].Type().GlobalType().Content().Kind())

	assert.Empty(t, imports[2].Module())
	assert.Empty(t, *imports[2].Name())
	assert.NotNil(t, imports[2].Type().TableType())
	assert.Equal(t, KindFuncref, imports[2].Type().TableType().Element().Kind())

	assert.Empty(t, imports[3].Module())
	assert.Empty(t, *imports[3].Name())
	assert.NotNil(t, imports[3].Type().MemoryType())
	assert.Equal(t, uint64(1), imports[3].Type().MemoryType().Minimum())
}

func TestModuleExports(t *testing.T) {
	wasm, err := Wat2Wasm(`
          (module
            (func (export "f"))
            (global (export "g") i32 (i32.const 0))
            (table (export "t") 1 funcref)
            (memory (export "m") 1)
          )
        `)
	assert.NoError(t, err)
	module, err := NewModule(NewEngine(), wasm)
	assert.NoError(t, err)

	exports := module.Exports()
	assert.Len(t, exports, 4)
	assert.Equal(t, "f", exports[0].Name())
	assert.NotNil(t, exports[0].Type().FuncType())
	assert.Len(t, exports[0].Type().FuncType().Params(), 0)
	assert.Len(t, exports[0].Type().FuncType().Results(), 0)

	assert.Equal(t, "g", exports[1].Name())
	assert.NotNil(t, exports[1].Type().GlobalType())
	assert.Equal(t, KindI32, exports[1].Type().GlobalType().Content().Kind())

	assert.Equal(t, "t", exports[2].Name())
	assert.NotNil(t, exports[2].Type().TableType())
	assert.Equal(t, KindFuncref, exports[2].Type().TableType().Element().Kind())

	assert.Equal(t, "m", exports[3].Name())
	assert.NotNil(t, exports[3].Type().MemoryType())
	assert.Equal(t, uint64(1), exports[3].Type().MemoryType().Minimum())
}

func TestModuleSerialize(t *testing.T) {
	engine := NewEngine()
	wasm, err := Wat2Wasm(`
          (module
            (func (export "f"))
            (global (export "g") i32 (i32.const 0))
            (table (export "t") 1 funcref)
            (memory (export "m") 1)
          )
        `)
	assert.NoError(t, err)
	module, err := NewModule(engine, wasm)
	assert.NoError(t, err)
	bytes, err := module.Serialize()
	assert.NoError(t, err)

	_, err = NewModuleDeserialize(engine, bytes)
	assert.NoError(t, err)

	tmpfile, err := ioutil.TempFile("", "example")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(bytes)
	assert.NoError(t, err)
	tmpfile.Close()
	_, err = NewModuleDeserializeFile(engine, tmpfile.Name())
	assert.NoError(t, err)
}
