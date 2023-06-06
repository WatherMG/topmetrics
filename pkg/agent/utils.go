package agent

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"topmetrics/pkg/metric"

	"github.com/shirou/gopsutil/v3/process"
)

// Send collecting metrics and send to server with defined interval
func Send(ctx context.Context, ch chan []*process.Process, duration *time.Duration, cfg *Config) error {
	ticker := time.NewTicker(*duration)
	defer ticker.Stop()
	var failedAttempts = 0

	for {
		if failedAttempts >= maxFailedAttempts {
			return fmt.Errorf("server is not responding after %d attempts", maxFailedAttempts)
		}
		select {
		case <-ctx.Done():
			log.Println("Metrics sending complete.")
			return nil
		case processes := <-ch:
			if len(processes) < cfg.MetricCount {
				log.Printf("Processes is < metricCount=%d, continue\n", cfg.MetricCount)
				continue
			}

			metricInfo, err := metric.NewMetric(ctx, processes, cfg.MetricCount)
			if err != nil {
				return err
			}

			dataType := getDataType(cfg.SerializeType)
			serializer, err := metric.NewSerializer(dataType)
			if err != nil {
				return err
			}

			data, err := metric.Marshal(serializer, metricInfo)
			if err != nil {
				return err
			}
			// add metadata to the transferred data, which the server will process and use the desired serializer
			data = append([]byte{dataType}, data...)

			if err := sendMetrics(data, cfg.Host, cfg.Port); err != nil {
				log.Println(err)
				failedAttempts++
			}
		}
	}
}

// sendMetrics used for send metrics to server
func sendMetrics(data []byte, host, port string) error {
	hostname := net.JoinHostPort(host, port)
	log.Printf("Try to connect to %s", hostname)

	conn, err := net.DialTimeout("tcp", hostname, sendTimeout)
	if err != nil {
		return fmt.Errorf("connecting error: %v", err.Error())
	}
	defer conn.Close()
	log.Printf("Connected to: %s", conn.RemoteAddr().String())

	n, err := conn.Write(data)
	if err != nil {
		return fmt.Errorf("send data error: %v", err.Error())
	}
	log.Printf("Data sent: %d bytes\n", n)

	return nil
}

// getDatatype used for defining in which type send data
func getDataType(serializerType string) (dataType byte) {
	switch strings.ToLower(serializerType) {
	case "j", "json":
		return metric.JSONType
	case "g", "gob":
		return metric.GOBType
	case "p", "proto":
		return metric.ProtoType
	default:
		log.Println("Use default serializer: Proto")
		return metric.ProtoType
	}
}
