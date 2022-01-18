package main

import (
	"order-confirmation-service/internal/api"
	"os"
)

const (
	portEnv     = "PORT"
	defaultPort = "8080"
)

func main() {
	port := os.Getenv(portEnv)
	if len(port) == 0 {
		port = defaultPort
	}
	api.NewServer(port).Serve()
}
