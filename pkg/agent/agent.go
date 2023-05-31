package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"topmetrics/pkg/metric"

	"github.com/shirou/gopsutil/v3/process"
)

func Send(ctx context.Context, ch chan []*process.Process, duration *time.Duration, metricCount *int, host, port string) {
	ticker := time.NewTicker(*duration)
	defer ticker.Stop()
	var failedAttempts = 0

	for {
		if failedAttempts >= 3 {
			log.Println("Server is not responding. Exit.")
			break
		}
		select {
		case <-ctx.Done():
			log.Println("Metrics sending complete.")
			return
		case processes := <-ch:
			if len(processes) < *metricCount {
				log.Printf("Processes is < metricCount=%d, continue\n", *metricCount)
				continue
			}

			metricInfo := &metric.Metric{}
			if err := metricInfo.Get(ctx, processes, metricCount); err != nil {
				log.Println(err)
			}
			jsonData, err := json.Marshal(metricInfo)
			if err != nil {
				log.Println(err)
			}
			if err := sendMetrics(jsonData, host, port, 3*time.Second); err != nil {
				log.Println(err)
				failedAttempts++
			}
		}
	}
}

func sendMetrics(jsonData []byte, host, port string, timeout time.Duration) error {
	hostname := fmt.Sprintf("%s:%s", host, port)
	log.Println("Try to connect to", hostname)

	conn, err := net.DialTimeout("tcp", hostname, timeout)
	if err != nil {
		return fmt.Errorf("error connecting: %v", err.Error())
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Println(err)
		}
	}()
	log.Printf("Connected to: %s", conn.RemoteAddr().String())
	n, err := conn.Write(jsonData)
	if err != nil {
		return fmt.Errorf("error send data: %v", err.Error())
	}
	log.Printf("Data sent: %d bytes\n", n)

	return nil
}
