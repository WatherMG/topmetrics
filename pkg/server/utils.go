package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"topmetrics/pkg/metric"
)

func closeConnection(conn net.Conn) {
	err := conn.Close()
	if err != nil {
		log.Printf("Connection error: can't close connection: %v\n", err)
	}
	log.Printf("Agent %s disconnected\n", conn.RemoteAddr())
}

func readData(conn io.Reader) (byte, []byte, error) {
	// dataType. Defining the type of data from client
	dataType := make([]byte, 1)
	_, err := conn.Read(dataType)
	if err != nil {
		return 0, nil, fmt.Errorf("can't read data type: %w", err)
	}

	var buf = make([]byte, maxReadBufferSize)
	n, err := conn.Read(buf)
	if err != nil {
		return 0, nil, fmt.Errorf("can't read data: %w", err)
	}
	log.Printf("Recived: %d bytes\n", n)
	return dataType[0], buf[:n], nil
}

func unmarshalData(dataType byte, data []byte) (*metric.Metric, error) {
	metricInfo := &metric.Metric{}

	serializer, err := metric.NewSerializer(dataType)
	if err != nil {
		return nil, fmt.Errorf("can't create serializer: %w", err)
	}

	log.Printf("Codec: %s", serializer)

	if err := metric.Unmarshal(serializer, data, metricInfo); err != nil {
		return nil, fmt.Errorf("can't unmarshal data: %w", err)
	}
	return metricInfo, nil
}

func logfileWriter(ch <-chan *metric.Metric) {
	metricInfo := <-ch
	log.Printf("Writing data")
	host := metricInfo.Hostname + "_" + metricInfo.HostId

	filePath := getFilePath(host)
	file, err := openFile(filePath)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	logMsg := metricInfo.BuildLogMsg()
	if err := fileWriter(file, []byte(logMsg)); err != nil {
		log.Println(err)
		return
	}

	log.Println("Data added into:", filePath)
}

func openFile(filePath string) (*os.File, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// filePathCache used for reduce i/o bound and hdd load
var filePathCache = make(map[string]string)
var filePathCacheMutex sync.Mutex

func getFilePath(host string) string {
	filePathCacheMutex.Lock()
	defer filePathCacheMutex.Unlock()

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
	return filePath
}

func createFilePath(host string) string {
	return filepath.Join(getWorkingDir(), "logs", host+".log")
}

var fileMutex sync.Mutex

func fileWriter(file io.Writer, data []byte) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()
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
