package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"topmetrics/pkg/metric"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:4444")
	if err != nil {
		log.Fatalf("Error listening: %v", err.Error())
		return
	}

	defer listener.Close()

	log.Printf("Listening on: %s", listener.Addr().String())

	for {
		log.Printf("Waiting conetcion from agent")
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting: %v", err.Error())
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		_ = conn.Close()
		log.Printf("Agent %s disconnected\n", conn.RemoteAddr())
	}()
	ch := make(chan metric.Metric)

	go clientWriter(ch)

	data := bufio.NewScanner(conn)
	log.Printf("Agent connected: %s", conn.RemoteAddr())

	for data.Scan() {
		text := data.Text()
		if strings.HasPrefix(text, "{") && strings.HasSuffix(text, "}") {
			var processes metric.Metric
			if err := json.Unmarshal([]byte(text), &processes); err != nil {
				log.Printf("error unmarshaling json message: %v", err)
				continue
			}
			ch <- processes
		}
	}

}
func clientWriter(ch <-chan metric.Metric) {
	procs := <-ch
	log.Printf("Write data")
	fmt.Println(procs.Hostname, time.Since(procs.SentAt).String()+" ago:")
	for _, process := range procs.Processes {
		fmt.Printf("PID:%d\tNAME: %s\tCPU USAGE: %.2f%%\tMEMORY USAGE:%.2f MB\n", process.PID, process.Name, process.CPUPercent, process.Memory)
	}
}
