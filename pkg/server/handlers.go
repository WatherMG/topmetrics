package server

type MetricHandler interface {
	HandleMetric(dataType byte, data []byte) error
}

type LogstashMetricHandler struct {
	server *Server
}

func (h *LogstashMetricHandler) HandleMetric(dataType byte, data []byte) error {
	return h.server.LogstashSender(data, dataType)
}

type FileMetricHandler struct {
	server *Server
}

func (h *FileMetricHandler) HandleMetric(dataType byte, data []byte) error {
	metricInfo, err := unmarshalData(dataType, data)
	if err != nil {
		return err
	}
	go logfileWriter(h.server.metricsQueue)
	if metricInfo.Processes != nil && metricInfo.Hostname != "" && metricInfo.HostId != "" {
		h.server.metricsQueue <- metricInfo
	}
	return nil
}
