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

	// Attach the rate limit middleware for 10 req/min
	vs.Use(ratelimit.NewTimeLimiter(time.Minute, 10))

	// Target server to forward
	vs.Forward("http://httpbin.org")

	fmt.Printf("Server listening on port: %d\n", port)
	err := vs.Listen()
	if err != nil {
		fmt.Errorf("Error: %s\n", err)
	}
}
