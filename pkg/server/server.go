package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"topmetrics/pkg/logstash"
)

func (serv *Server) receiveMetricHandler(conn net.Conn) {
	defer closeConnection(conn)
	for {
		dataType, data, err := readData(conn)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("Connection closed by remote host")
				return
			}
			log.Println(err)
			return
		}

		handler, ok := serv.metricHandlers[dataType]
		if !ok {
			log.Printf("Unsupported data type: %d", dataType)
			continue
		}

		err = handler.HandleMetric(dataType, data)
		if err != nil {
			log.Println(err)
		}
	}
}

func (serv *Server) LogstashSender(data []byte, dataType byte) error {
	if serv.failedAttempts >= maxFailedAttempts {
		return fmt.Errorf("logstash is not responding after %d attempts. dataType flag: %d", maxFailedAttempts, dataType)
	}
	config, err := logstash.GetConfig(dataType)
	if err != nil {
		return err
	}
	/*defer config.Cancel()*/
	err = config.SendMetric(serv.ctx, data)
	if err != nil {
		if a := serv.ctx.Err(); a != nil {
			log.Println(a.Error())
		}
		serv.failedAttempts++
		return err
	}

	return nil
}
