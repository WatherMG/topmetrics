package metric

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

func Collect(ctx context.Context, ch chan []*process.Process, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Process information gathering complete.")
			return
		case <-ticker.C:
			processes, err := process.ProcessesWithContext(ctx)
			if err != nil {
				log.Println(err)
				continue
			}
			sortProcesses(ctx, processes)
			ch <- processes
		}
	}
}

func NewProcess(ctx context.Context, process *process.Process) (*ProcessInfo, error) {
	p := &ProcessInfo{}
	name, err := process.NameWithContext(ctx)
	if err != nil {
		return nil, err
	}
	cpu, err := process.CPUPercentWithContext(ctx)
	if err != nil {
		return nil, err
	}
	mem, err := process.MemoryInfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	p.Pid = process.Pid
	p.Name = name
	p.CpuPercent = cpu
	p.MemoryUsage = float64(mem.RSS) / math.Pow(1024, 2)

	return p, nil
}

func sortProcesses(ctx context.Context, processes []*process.Process) {
	sort.SliceStable(processes, func(i, j int) bool {
		n, err := processes[i].CPUPercentWithContext(ctx)
		if err != nil {
			return false
		}
		m, err := processes[j].CPUPercentWithContext(ctx)
		if err != nil {
			return false
		}
		return n > m
	})
}
