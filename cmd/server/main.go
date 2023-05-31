package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"topmetrics/pkg/metric"
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
	host := procs.HostID

	buff := &bytes.Buffer{}
	file, err := os.OpenFile("./logs/"+host+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}

	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer file.Close()

	for _, process := range procs.Processes {
		fmt.Fprintf(buff, "PID:%d\nNAME: %s\nCPU USAGE: %.2f%%\nMEMORY USAGE: %.2f MB\n", process.PID, process.Name, process.CPUPercent, process.Memory)
		fmt.Fprintf(buff, "--------------------------------------------------\n")
	}
	if _, err := file.Write(buff.Bytes()); err != nil {
		log.Println(err)
	}

	fullPath := filepath.Join(dir)
	log.Println("Data added into", fullPath)
}
