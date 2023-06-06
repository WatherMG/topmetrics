package metric

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/process"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	hostInfoFormat    = "HOSTINFO: %s %s %v "
	processInfoFormat = "PROCESSINFO: pid: %d process_name: %s cpu_percent: %.2f memory_usage: %.2f "
)

func NewMetric(ctx context.Context, processes []*process.Process, metricCount int) (*Metric, error) {
	m := &Metric{}
	m.Processes = make([]*ProcessInfo, metricCount)

	for i, v := range processes[:metricCount] {
		processInfo, err := NewProcess(ctx, v)
		if err != nil {
			return nil, err
		}
		m.Processes[i] = processInfo
	}

	hostname, err := host.Info()
	if err != nil {
		return nil, err
	}

	m.Hostname = hostname.Hostname
	m.HostId = hostname.HostID
	m.SentAt = timestamppb.Now()

	return m, nil
}

// BuildLogMsg used for create line for log file. We can't use String() method for the struct Metric because protobuf used it.
// String format from protobuf doesn't have clear delimiters between fields, and we can't use logstash filters.
func (m *Metric) BuildLogMsg() string {
	var buf strings.Builder
	buf.Grow(1024) // pre-allocate 1KB buffer
	fmt.Fprintf(&buf, hostInfoFormat, m.Hostname, m.HostId, m.SentAt.AsTime().Format(time.RFC3339Nano))
	for _, p := range m.Processes {
		fmt.Fprintf(&buf, processInfoFormat, p.Pid, p.Name, p.CpuPercent, p.MemoryUsage)
	}
	buf.WriteRune('\n')
	return buf.String()
}
