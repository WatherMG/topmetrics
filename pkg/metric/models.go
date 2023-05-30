package metric

import (
	"time"
)

type ProcessInfo struct {
	PID        int32   `json:"pid"`
	Name       string  `json:"name"`
	CPUPercent float64 `json:"cpu_percent"`
	Memory     float64 `json:"memory"`
}

type Metric struct {
	HostID    string         `json:"host_id"`
	Hostname  string         `json:"hostname"`
	SentAt    time.Time      `json:"sent_at"`
	Processes []*ProcessInfo `json:"processes"`
}
