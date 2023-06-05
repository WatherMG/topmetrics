package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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

	metricsCh := make(chan metric.Metric)

	go clientWriter(metricsCh)

	data := bufio.NewScanner(conn)
	log.Printf("Agent connected: %s", conn.RemoteAddr())

	for data.Scan() {
		text := data.Text()
		if strings.HasPrefix(text, "{") && strings.HasSuffix(text, "}") {
			var metric metric.Metric
			if err := json.Unmarshal([]byte(text), &metric); err != nil {
				log.Printf("error unmarshaling json message: %v", err)
				continue
			}
			if metric.Processes != nil && metric.Hostname != "" && metric.HostID != "" {
				metricsCh <- metric
			}
		}
	}
}

func clientWriter(ch <-chan metric.Metric) {
	metric := <-ch
	log.Printf("Write data")
	host := metric.Hostname + metric.HostID

	filePath, ok := filePathCache[host]
	if !ok {
		filePath = createFilePath(host)
		filePathCache[host] = filePath
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return
	}

	buff := &bytes.Buffer{}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file)

	fmt.Fprintf(buff, "HOSTINFO: %s %s %v ", metric.Hostname, metric.HostID, metric.SentAt.Format(time.RFC3339Nano))
	for _, process := range metric.Processes {
		pid := process.PID
		name := process.Name
		cpuUsage := process.CPUPercent
		memoryUsage := process.Memory
		if _, err := fmt.Fprintf(buff, "PROCESSINFO: pid: %d process_name: %s cpu_percent: %.2f memory_usage: %.2f ", pid, name, cpuUsage, memoryUsage); err != nil {
			log.Panic("writing error: ", err, process)
		}
	}
	buff.WriteRune('\n')
	if err := fileWriter(file, buff.Bytes()); err != nil {
		log.Println(err)
		return
	}

	log.Println("Data added into", filePath)
}

var filePathCache = make(map[string]string)

func createFilePath(host string) string {
	return filepath.Join(getWorkingDir(), "logs", host+".log")
}

func fileWriter(file io.Writer, data []byte) error {
	var mutex sync.Mutex
	mutex.Lock()
	defer mutex.Unlock()
	if _, err := file.Write(data); err != nil {
		return err
	}
	return nil
}

func getWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return dir
}
