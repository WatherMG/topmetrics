package logstash

import (
	"fmt"
	"log"

	"topmetrics/pkg/connector"
	"topmetrics/pkg/metric"
)

func init() {
	log.Println("[Logstash Logger] initializing...")
}

var logstashConfigMap = map[byte]*Logstash{
	metric.ProtoType: NewConfig("192.168.0.199", "5044", "tcp"),
	metric.JSONType:  NewConfig("192.168.0.199", "5045", "tcp"),
}

type Logstash struct {
	*connector.Connector
	/*Ctx    context.Context
	Cancel context.CancelFunc*/
}

// func (l *Logstash) SendMetric(data []byte) error {
// 	if err := l.initConnection(); err != nil {
// 		return fmt.Errorf("logstash: initialize error: %w", err)
// 	}
// 	defer l.Close()
//
// 	numBytes, err := l.Write(data)
// 	if err != nil {
// 		return fmt.Errorf("logstash: failed to send log message: %w", err)
// 	}
// 	log.Printf("Sent: %d bytes\n", numBytes)
// 	return nil
// }
//
// func (l *Logstash) initConnection() error {
// 	config := connector.NewConnector(l.host, l.port, l.connectionType)
// 	conn, err := config.Connect()
// 	if err != nil {
// 		return err
// 	}
// 	l.Conn = conn
// 	return nil
// }

func NewConfig(host string, port string, connectionType string) *Logstash {
	conn := connector.NewConnector(host, port, connectionType)

	return &Logstash{conn}
}

func GetConfig(dataType byte) (*Logstash, error) {
	if config, ok := logstashConfigMap[dataType]; ok {
		return config, nil
	}
	return nil, fmt.Errorf("logstash: unknown logstash config: flag: %b", dataType)
}
