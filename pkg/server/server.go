package server

import (
	"fmt"
	"log"
	"net"

	"topmetrics/pkg/logstash"
	"topmetrics/pkg/metric"
)

const (
	maxReadBufferSize = 1024
	maxCacheSize      = 100
	maxFailedAttempts = 3
)

func ReceiveMetricHandler(conn net.Conn) {
	metricsCh := make(chan *metric.Metric)
	dataType, data, err := readData(conn)
	if err != nil {
		log.Println(err)
	}

	defer closeConnection(conn)

	switch dataType {
	case metric.JSONType, metric.ProtoType:
		err = sendDataToLogstash(data, dataType)
		if err != nil {
			log.Println(err)
		}
	case metric.GOBType:
		metricInfo, err := unmarshalData(dataType, data)
		if err != nil {
			log.Println(err)
		}
		go logfileWriter(metricsCh)

		if metricInfo.Processes != nil && metricInfo.Hostname != "" && metricInfo.HostId != "" {
			metricsCh <- metricInfo
		}
	}
}

var failedAttempts = 0

func sendDataToLogstash(data []byte, dataType byte) error {
	if failedAttempts >= maxFailedAttempts {
		return fmt.Errorf("logstash is not responding after %d attempts. dataType flag: %d", maxFailedAttempts, dataType)
	}
	config, err := logstash.GetLogstashConfig(dataType)
	if err != nil {
		return err
	}
	err = config.SendMetric(data)
	if err != nil {
		failedAttempts++
		return err
	}
	return nil
}
