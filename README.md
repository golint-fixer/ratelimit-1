# ratelimit [![Build Status](https://travis-ci.org/vinxi/ratelimit.png)](https://travis-ci.org/vinxi/ratelimit) [![GoDoc](https://godoc.org/github.com/vinxi/ratelimit?status.svg)](https://godoc.org/github.com/vinxi/ratelimit) [![Coverage Status](https://coveralls.io/repos/github/vinxi/ratelimit/badge.svg?branch=master)](https://coveralls.io/github/vinxi/ratelimit?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/vinxi/ratelimit)](https://goreportcard.com/report/github.com/vinxi/ratelimit)

Efficient token bucket implementation rate limiter for your proxies.

## Installation

```bash
go get -u gopkg.in/vinxi/ratelimit.v0
```

## API

See [godoc](https://godoc.org/github.com/vinxi/ratelimit) reference.

## Example

#### Default log to stdout

```go
package main

import (
  "fmt"
  "gopkg.in/vinxi/ratelimit.v0"
  "gopkg.in/vinxi/vinxi.v0"
)

const port = 3100

func main() {
  // Create a new vinxi proxy
  vs := vinxi.NewServer(vinxi.ServerOptions{Port: port})
  
  // Attach the log middleware 
  vs.Use(log.Default)
  
  // Target server to forward
  vs.Forward("http://httpbin.org")

  fmt.Printf("Server listening on port: %d\n", port)
  err := vs.Listen()
  if err != nil {
    fmt.Errorf("Error: %s\n", err)
  }
}
```

## License

MIT
