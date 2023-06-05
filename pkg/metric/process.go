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

func Collect(ctx context.Context, ch chan []*process.Process, duration *time.Duration) {
	ticker := time.NewTicker(*duration)
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

func (p *ProcessInfo) Get(ctx context.Context, process *process.Process) (*ProcessInfo, error) {
	name, err := process.NameWithContext(ctx)
	if err != nil {
		return p, err
	}
	cpu, err := process.CPUPercentWithContext(ctx)
	if err != nil {
		return p, err
	}
	mem, err := process.MemoryInfoWithContext(ctx)
	if err != nil {
		return p, err
	}

	p.Pid = process.Pid
	p.Name = name
	p.CpuPercent = cpu
	p.MemoryUsage = float64(mem.RSS) / math.Pow(1024, 2)

	return p, nil
}

// func (p *ProcessInfo) String() string {
// 	return fmt.Sprintf("%d, %s, %.2f%%, %.2fMB\n", p.PID, p.Name, p.CPUPercent, p.Memory)
// }

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
