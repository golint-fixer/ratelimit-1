# ratelimit [![Build Status](https://travis-ci.org/vinxi/ratelimit.png)](https://travis-ci.org/vinxi/ratelimit) [![GoDoc](https://godoc.org/github.com/vinxi/ratelimit?status.svg)](https://godoc.org/github.com/vinxi/ratelimit) [![Coverage Status](https://coveralls.io/repos/github/vinxi/ratelimit/badge.svg?branch=master)](https://coveralls.io/github/vinxi/ratelimit?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/vinxi/ratelimit)](https://goreportcard.com/report/github.com/vinxi/ratelimit)

Simple, efficient token bucket rate limiter for your proxies.

Supports filters and whitelist exceptions to determine when to apply the rate limiter.

## Installation

```bash
go get -u gopkg.in/vinxi/ratelimit.v0
```

## API

See [godoc](https://godoc.org/github.com/vinxi/ratelimit) reference.

## Example

#### Rate limit based on time window

```go
package main

import (
  "fmt"
  "time"
  "gopkg.in/vinxi/ratelimit.v0"
  "gopkg.in/vinxi/vinxi.v0"
)

const port = 3100

func main() {
  // Create a new vinxi proxy
  vs := vinxi.NewServer(vinxi.ServerOptions{Port: port})
  
  // Attach the rate limit middleware for 100 req/min
  vs.Use(ratelimit.NewTimeLimiter(time.Minute, 100))
  
  // Target server to forward
  vs.Forward("http://httpbin.org")

  fmt.Printf("Server listening on port: %d\n", port)
  err := vs.Listen()
  if err != nil {
    fmt.Errorf("Error: %s\n", err)
  }
}
```

#### Limit requests per second

```go
package main

import (
  "fmt"
  "gopkg.in/vinxi/ratelimit.v0"
  "gopkg.in/vinxi/vinxi.v0"
  "time"
)

const port = 3100

func main() {
  // Create a new vinxi proxy
  vs := vinxi.NewServer(vinxi.ServerOptions{Port: port})

  // Attach the ratelimit middleware for 10 req/second
  vs.Use(ratelimit.NewRateLimiter(10, 10))

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
