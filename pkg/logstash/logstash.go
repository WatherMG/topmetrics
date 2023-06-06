package logstash

import (
	"fmt"
	"log"
	"net"
	"time"

	"topmetrics/pkg/metric"
)

func init() {
	log.Println("[Logstash Logger] initializing...")
}

const sendTimeout = 15 * time.Second

var logstashConfigMap = map[byte]*Logstash{
	metric.ProtoType: NewConfig("192.168.0.199", "5044", "tcp", sendTimeout),
	metric.JSONType:  NewConfig("192.168.0.199", "5045", "tcp", sendTimeout),
}

type Logstash struct {
	host, port, connectionType string
	timeout                    time.Duration
	net.Conn
}

func (l *Logstash) SendMetric(data []byte) error {
	if err := l.initConnection(); err != nil {
		return fmt.Errorf("logstash: initialize error: %w", err)
	}
	defer l.Close()

	numBytes, err := l.Write(data)
	if err != nil {
		return fmt.Errorf("logstash: failed to send log message: %w", err)
	}
	log.Printf("Sent: %d bytes\n", numBytes)
	return nil
}

func (l *Logstash) setTimeout() error {
	deadline := time.Now().Add(l.timeout)
	err := l.SetDeadline(deadline)
	if err != nil {
		return err
	}
	return nil
}

func (l *Logstash) initConnection() error {
	host := net.JoinHostPort(l.host, l.port)
	conn, err := net.Dial(l.connectionType, host)
	if err != nil {
		return fmt.Errorf("connection error: failed to connect Logstash: %w", err)
	}
	l.Conn = conn
	if err = l.setTimeout(); err != nil {
		return fmt.Errorf("set timeout error: %w", err)
	}
	return nil
}

func NewConfig(host string, port string, connectionType string, timeout time.Duration) *Logstash {
	return &Logstash{host: host, port: port, connectionType: connectionType, timeout: timeout}
}

func GetLogstashConfig(dataType byte) (*Logstash, error) {
	if config, ok := logstashConfigMap[dataType]; ok {
		return config, nil
	}
	return nil, fmt.Errorf("logstash: unknown logstash config: flag: %b", dataType)
}
