package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"topmetrics/pkg/metric"

	"github.com/shirou/gopsutil/v3/process"
)

// HandleMetric collecting metrics and send to server with defined interval
func (agnt *Agent) HandleMetric(ch chan []*process.Process) error {
	for {
		if agnt.Config.FailAttempts >= maxFailedAttempts {
			return fmt.Errorf("server at %s:%s is not responding after %d attempts", agnt.Connector.Host, agnt.Connector.Port, maxFailedAttempts)
		}
		select {
		case <-agnt.Ctx.Done():
			log.Println("Metrics sending complete.")
			err := agnt.Connector.Close()
			if err != nil {
				return err
			}
			close(ch)
			return nil
		case processes := <-ch:
			if len(processes) < agnt.Config.Count {
				log.Printf("System has %d processes. Current metric count=%d\n", len(processes), agnt.Config.Count)
				continue
			}
			metricInfo, err := metric.NewMetric(agnt.Ctx, processes, agnt.Config.Count)
			if err != nil {
				return err
			}
			metricInfo.Hostname = agnt.Hostname
			if err = agnt.sendMetrics(agnt.Ctx, metricInfo); err != nil {
				log.Println(err)
				log.Println("Try to reconnect...")
				time.Sleep(reconnectTimeout)
				agnt.Config.FailAttempts++
				continue
			}
			agnt.Config.FailAttempts = 0
		}
	}
}

func (agnt *Agent) sendMetrics(ctx context.Context, metricInfo *metric.Metric) error {
	dataType := agnt.getDataType()
	serializer, err := metric.NewSerializer(dataType)
	if err != nil {
		return err
	}
	data, err := metric.Marshal(serializer, metricInfo)
	if err != nil {
		return err
	}
	// add metadata to the transferred data, which the server will process and use the desired serializer
	data = append([]byte{dataType}, data...)
	err = agnt.Connector.SendMetric(ctx, data)
	return err
}
