package main

import (
	"log"

	"topmetrics/pkg/agent"
	"topmetrics/pkg/metric"

	"github.com/shirou/gopsutil/v3/process"
)

func main() {
	agnt, err := agent.NewConfig()
	if err != nil {
		log.Fatalf("Agent error: %v", err)
	}
	defer agnt.Cancel()

	processes := make(chan []*process.Process)
	go metric.Collect(agnt.Ctx, processes, agnt.Config.Interval)

	if err = agnt.HandleMetric(processes); err != nil {
		log.Printf("send error: %v\n", err)
	}
}
