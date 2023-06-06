package main

import (
	"context"
	"flag"
	"log"
	"time"

	"topmetrics/pkg/agent"
	"topmetrics/pkg/metric"

	"github.com/shirou/gopsutil/v3/process"
)

var (
	metricCount   = flag.Int("count", 5, "Number of metrics to send")
	interval      = flag.Duration("interval", 5*time.Second, "Sending interval in sec")
	timeout       = flag.Duration("timeout", 1*time.Minute, "Duration of metrics sending in minutes")
	host          = flag.String("host", "192.168.0.199", "Server address")
	port          = flag.String("port", "8080", "Server port")
	hostname      = flag.String("hostname", "", "Custom hostname")
	serializeType = flag.String("type", "proto", "Type of serialization")
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	processes := make(chan []*process.Process)
	config, err := agent.NewConfig(*host, *port, *hostname, *serializeType, *metricCount)
	if err != nil {
		log.Printf("Config error: %v", err)
	}

	go metric.Collect(ctx, processes, interval)

	if err = agent.Send(ctx, processes, interval, config); err != nil {
		log.Printf("Send error: %v", err)
	}
}
