<div align="center">
  <h1><code>wasmtime-go</code></h1>

  <p>
    <strong>Go embedding of
    <a href="https://github.com/bytecodealliance/wasmtime">Wasmtime</a></strong>
  </p>

  <strong>A <a href="https://bytecodealliance.org/">Bytecode Alliance</a> project</strong>

  <p>
    <a href="https://pkg.go.dev/github.com/bytecodealliance/wasmtime-go/pkg/wasmtime">
      Documentation
    </a>
  </p>

</div>

## Usage

You can import this extension directly from GitHub:

```go
import "github.com/bytecodealliance/wasmtime-go/pkg/wasmtime"
```

This extension uses [cgo](https://golang.org/cmd/cgo/) to use the Wasmtime C
API. Currently only x86\_64 Windows, macOS, and Linux are supported. You'll need
to arrange to have Wasmtime installed on your system, for example by downloading
the C API [from the releases
page](https://github.com/bytecodealliance/wasmtime/releases) and adjusting
`CGO_CFLAGS` with `-I` to the `include` directory and `CGO_LDFLAGS` with `-L` to
the `lib` directory.

## Contributing

So far this extension has been written by folks who are primarily Rust
programmers, so it's highly likely that there's some faux pas in terms of Go
idioms. Feel free to send a PR to help make things more idiomatic if you see
something!

To work on this extension locally you'll first want to clone the project:

```sh
$ git clone https://github.com/bytecodealliance/wasmtime-go
```

Next up you'll want to have a [local Wasmtime build
available](https://bytecodealliance.github.io/wasmtime/contributing-building.html).
Once you've got that you can work locally on this library with:

```
$ WASMTIME=/path/to/wasmtime ./ci/run-local.sh go test ./...
```

The `run-local.sh` script will set up necessary cgo environment variables to
link against wasmtime.

And after that you should be good to go!
