package metric

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/process"
)

func (m *Metric) Get(ctx context.Context, processes []*process.Process, metricCount *int) error {
	m.Processes = make([]*ProcessInfo, *metricCount)

	for i, v := range processes[:*metricCount] {
		procInfo := &ProcessInfo{}
		info, err := procInfo.Get(ctx, v)
		if err != nil {
			return err
		}
		m.Processes[i] = info
	}

	hostname, err := host.Info()
	if err != nil {
		return err
	}

	m.Hostname = hostname.Hostname
	m.HostID = hostname.HostID
	m.SentAt = time.Now()

	return nil
}
