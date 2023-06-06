package agent

import (
	"os"
	"time"
)

const (
	maxFailedAttempts = 3
	sendTimeout       = 3 * time.Second
)

type Config struct {
	Host          string
	Port          string
	Hostname      string
	SerializeType string
	MetricCount   int
}

func NewConfig(host, port, hostname, serializeType string, metricCount int) (*Config, error) {
	config := &Config{
		Host:          host,
		Port:          port,
		SerializeType: serializeType,
		MetricCount:   metricCount,
	}
	if err := config.SetHostname(hostname); err != nil {
		return nil, err
	}
	return config, nil
}

func (c *Config) SetHostname(hostname string) error {
	if hostname != "" {
		c.Hostname = hostname
		return nil
	}
	name, err := os.Hostname()
	if err != nil {
		return err
	}
	c.Hostname = name
	return nil
}
