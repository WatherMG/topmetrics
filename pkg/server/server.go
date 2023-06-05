package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"topmetrics/pkg/metric"
)

const (
	maxReadBufferSize  = 1024
	maxCacheSize       = 100
)

func HandleConnection(conn net.Conn) {
	// dataType. Defining the type of data from client
	dataType := make([]byte, 1)
	_, err := conn.Read(dataType)
	if err != nil {
		log.Printf("Can't read data type: %v\n", err)
		return
	}

	defer func() {
		err = conn.Close()
		if err != nil {
			log.Printf("Can't close connection: %v\n", err)
		}
		log.Printf("Agent %s disconnected\n", conn.RemoteAddr())
	}()

	metrica := &metric.Metric{}
	var buf = make([]byte, maxReadBufferSize)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Recived: %d bytes\n", n)
	data := buf[:n]

	metricsCh := make(chan *metric.Metric)
	go agentWriter(metricsCh)

	serializer, err := createSerializer(dataType[0])
	if err != nil {
		log.Println(err)
	}

	if err := metric.Unmarshal(serializer, data, metrica); err != nil {
		log.Println(err)
		return
	}

	if metrica.Processes != nil && metrica.Hostname != "" && metrica.HostId != "" {
		metricsCh <- metrica
	}
}

// createSerializer create serializer depending on dataType
func createSerializer(dataType byte) (interface{}, error) {
	var serializer interface{}
	switch dataType {
	case metric.JSONType:
		serializer = &metric.JSONSerialize{}
	case metric.GOBType:
		serializer = &metric.GOBSerialize{}
	case metric.ProtoType:
		serializer = &metric.ProtoSerialize{}
	default:
		return nil, fmt.Errorf("unknown data type: %v", dataType)
	}
	return serializer, nil
}

func agentWriter(ch <-chan *metric.Metric) {
	metrica := <-ch
	log.Printf("Writing data")
	host := metrica.Hostname + "_" + metrica.HostId

	filePath, ok := filePathCache[host]
	if !ok {
		filePath = createFilePath(host)
		// remove key from filePathCache if buffer >= maxCacheSize
		if len(filePathCache) >= maxCacheSize {
			for k := range filePathCache {
				delete(filePathCache, k)
				break
			}
		}
		filePathCache[host] = filePath
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file)

	sentAt := metrica.SentAt.AsTime()
	buff := &bytes.Buffer{}

	fmt.Fprintf(buff, "HOSTINFO: %s %s %v ", metrica.Hostname, metrica.HostId, sentAt.Format(time.RFC3339Nano))
	for _, process := range metrica.Processes {
		pid := process.Pid
		name := process.Name
		cpuUsage := process.CpuPercent
		memoryUsage := process.MemoryUsage
		if _, err := fmt.Fprintf(buff, "PROCESSINFO: pid: %d process_name: %s cpu_percent: %.2f memory_usage: %.2f ", pid, name, cpuUsage, memoryUsage); err != nil {
			log.Panic("writing error: ", err, process)
		}
	}
	buff.WriteRune('\n')
	if err := fileWriter(file, buff.Bytes()); err != nil {
		log.Println(err)
		return
	}

	log.Println("Data added into:", filePath)
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
