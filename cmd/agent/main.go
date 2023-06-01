package main

import (
	"context"
	"flag"
	"time"

	"topmetrics/pkg/agent"
	"topmetrics/pkg/metric"

	"github.com/shirou/gopsutil/v3/process"
)

var (
	metricCount = flag.Int("count", 5, "Number of metrics to send")
	interval    = flag.Duration("interval", 5*time.Second, "Sending interval in sec")
	timeout     = flag.Duration("timeout", 1*time.Minute, "Duration of metrics sending in minutes")
	host        = flag.String("host", "192.168.0.199", "Server address")
	port        = flag.String("port", "8080", "Server port")
	hostname    = flag.String("hostname", "", "Custom hostname")
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	processes := make(chan []*process.Process)

	go metric.Collect(ctx, processes, interval)

	agent.Send(ctx, processes, interval, metricCount, *host, *port, *hostname)
}
