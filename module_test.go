package wasmtime

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModule(t *testing.T) {
	_, err := NewModule(NewEngine(), []byte{})
	require.Error(t, err)
	_, err = NewModule(NewEngine(), []byte{1})
	require.Error(t, err)
}

func TestModuleValidate(t *testing.T) {
	require.NotNil(t, ModuleValidate(NewEngine(), []byte{}), "expected an error")
	require.NotNil(t, ModuleValidate(NewEngine(), []byte{1}), "expected an error")
	wasm, err := Wat2Wasm(`(module)`)
	require.NoError(t, err)
	require.Nil(t, ModuleValidate(NewEngine(), wasm), "expected valid module")
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
	require.NoError(t, err)
	module, err := NewModule(NewEngine(), wasm)
	require.NoError(t, err)
	imports := module.Imports()
	require.Len(t, imports, 4)
	require.Equal(t, "", imports[0].Module())
	require.Equal(t, "f", *imports[0].Name())
	require.NotNil(t, imports[0].Type().FuncType())
	require.Len(t, imports[0].Type().FuncType().Params(), 0)
	require.Len(t, imports[0].Type().FuncType().Results(), 0)

	require.Equal(t, "a", imports[1].Module())
	require.Equal(t, "g", *imports[1].Name())
	require.NotNil(t, imports[1].Type().GlobalType())
	require.Equal(t, KindI32, imports[1].Type().GlobalType().Content().Kind())

	require.Empty(t, imports[2].Module())
	require.Empty(t, *imports[2].Name())
	require.NotNil(t, imports[2].Type().TableType())
	require.Equal(t, KindFuncref, imports[2].Type().TableType().Element().Kind())

	require.Empty(t, imports[3].Module())
	require.Empty(t, *imports[3].Name())
	require.NotNil(t, imports[3].Type().MemoryType())
	require.Equal(t, uint64(1), imports[3].Type().MemoryType().Minimum())
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
	require.NoError(t, err)
	module, err := NewModule(NewEngine(), wasm)
	require.NoError(t, err)

	exports := module.Exports()
	require.Len(t, exports, 4)
	require.Equal(t, "f", exports[0].Name())
	require.NotNil(t, exports[0].Type().FuncType())
	require.Len(t, exports[0].Type().FuncType().Params(), 0)
	require.Len(t, exports[0].Type().FuncType().Results(), 0)

	require.Equal(t, "g", exports[1].Name())
	require.NotNil(t, exports[1].Type().GlobalType())
	require.Equal(t, KindI32, exports[1].Type().GlobalType().Content().Kind())

	require.Equal(t, "t", exports[2].Name())
	require.NotNil(t, exports[2].Type().TableType())
	require.Equal(t, KindFuncref, exports[2].Type().TableType().Element().Kind())

	require.Equal(t, "m", exports[3].Name())
	require.NotNil(t, exports[3].Type().MemoryType())
	require.Equal(t, uint64(1), exports[3].Type().MemoryType().Minimum())
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
	require.NoError(t, err)
	module, err := NewModule(engine, wasm)
	require.NoError(t, err)
	bytes, err := module.Serialize()
	require.NoError(t, err)

	_, err = NewModuleDeserialize(engine, bytes)
	require.NoError(t, err)

	tmpfile, err := ioutil.TempFile("", "example")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(bytes)
	require.NoError(t, err)
	tmpfile.Close()
	_, err = NewModuleDeserializeFile(engine, tmpfile.Name())
	require.NoError(t, err)
}
