package main

import (
	"log"

	"forum/internal/server"
)

// main starts the forum server and terminates the application
// if the server fails to start or encounters a fatal error.
func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}