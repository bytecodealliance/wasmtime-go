#!/bin/bash

# This script tests the usage of wasmtime-go in a Go project
# that vendors its dependencies. This is used to check if
# all the header files and binaries are well copied in the
# vendor directory.

# Create a test directory
rm -rf test-vendoring-project
mkdir test-vendoring-project && cd test-vendoring-project || return 1

# Initialize a Go project
go mod init "test-vendoring-project"

# Add dependency to wasmtime-go
cat << E0F >> go.mod
require "github.com/bytecodealliance/wasmtime-go" v0.0.0
replace "github.com/bytecodealliance/wasmtime-go" v0.0.0 => "../"
E0F

# Add a basic test which uses wasmtime-go
cat << "E0F" >> main_test.go
package main

import (
	"testing"

	"github.com/bytecodealliance/wasmtime-go"
)

func TestMain(t *testing.T) {
	wasmtime.NewStore(wasmtime.NewEngine())
}
E0F

# Vendor dependency
go mod vendor

# Then execute the test
go test .

echo "Success"