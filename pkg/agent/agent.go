package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"topmetrics/pkg/metric"

	"github.com/shirou/gopsutil/v3/process"
)

func Send(ctx context.Context, ch chan []*process.Process, duration *time.Duration, metricCount *int, host, port string) {
	ticker := time.NewTicker(*duration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Metrics sending complete.")
			return
		case processes := <-ch:
			if len(processes) < *metricCount {
				continue
			}

			metricInfo := &metric.Metric{}
			if err := metricInfo.Get(ctx, processes, metricCount); err != nil {
				log.Println(err)
				continue
			}

			jsonData, err := json.Marshal(metricInfo)
			if err != nil {
				log.Println(err)
				continue
			}
			if host == "" || port == "" {
				log.Println("Host or port is missing.")
				os.Exit(1)
			}
			if err := sendMetrics(jsonData, host, port); err != nil {
				log.Println(err)
				continue
			}
		}
	}
}

func sendMetrics(jsonData []byte, host, port string) error {
	conn, err := net.Dial("tcp", host+":"+port)
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
