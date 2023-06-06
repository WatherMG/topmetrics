package main

import (
	"log"
	"net"

	"topmetrics/pkg/server"
)

func main() {
	tcp := &net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 8080,
	}
	listener, err := net.ListenTCP("tcp", tcp)
	if err != nil {
		log.Fatalf("Error listening: %v", err.Error())
		return
	}

	defer listener.Close()

	log.Printf("Listening on: %s", listener.Addr().String())

	for {
		log.Printf("Waiting connection from agent")
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting: %v", err.Error())
			continue
		}
		go server.ReceiveMetricHandler(conn)
	}
}
