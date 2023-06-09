package main

import (
	"log"

	"topmetrics/pkg/server"
)

func main() {
	newServer, err := server.NewServer("0.0.0.0:8080", "tcp")
	if err != nil {
		log.Fatalf("Error creating server: %v", err)
	}
	newServer.Run()
}
