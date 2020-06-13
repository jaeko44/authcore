package main

import (
	"authcore.io/authcore/internal/server"
)

// Initialize and start the server.
func main() {
	server := server.NewServer()
	server.Start()
}
