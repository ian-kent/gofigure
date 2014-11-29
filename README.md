gofigure  [![GoDoc](https://godoc.org/github.com/ian-kent/gofigure?status.svg)](https://godoc.org/github.com/ian-kent/gofigure)
========

Go configuration made easy!

- Just define a struct and call Gofigure
- Supports strings, ints/uints/floats, slices and nested structs
- Supports environment variables and command line flags

### Example

`go get github.com/ian-kent/gofigure`

```go
package main

import "github.com/ian-kent/gofigure"

type config struct {
  gofigure interface{} `envPrefix:"BAR" order:"flag,env"`
  RemoteAddr string `env:"REMOTE_ADDR" flag:"remote-addr" flagDesc:"Remote address"`
  LocalAddr  string `env:"LOCAL_ADDR" flag:"local-addr" flagDesc:"Local address"`
  NumCPU int `env:"NUM_CPU" flag:"num-cpu" flagDesc:"Number of CPUs"`
  Sources []string `env:"SOURCES" flag:"source" flagDesc:"Source URL (can be provided multiple times)"`
  Numbers []int `env:"NUMBERS" flag:"number" flagDesc:"Number (can be provided multiple times)"`
  Advanced struct{
      MaxBytes int64 `env:"MAX_BYTES" flag:"max-bytes" flagDesc:"Max bytes"`
      MaxErrors int64  `env:"MAX_ERRORS" flag:"max-errors" flagDesc:"Max errors"`
  }
}

func main() {
  var cfg config
  err := gofigure.Gofigure(&cfg)
  if err != nil {
    log.Fatal(err)
  }
  // use cfg
}
```

### gofigure field

The gofigure field is used to configure Gofigure.

The `order` tag is used to set configuration source order, e.g.
environment variables first then command line options second.

Any field matching `camelCase` format will be parsed into `camel`
and `case`, and passed to the source matching `camel`.

For example, the `envPrefix` field is split into `env` and `prefix`,
and the tag value is passed to the environment variable source as
the `prefix` parameter.

### Licence

Copyright ©‎ 2014, Ian Kent (http://www.iankent.eu).

Released under MIT license, see [LICENSE](LICENSE.md) for details.
