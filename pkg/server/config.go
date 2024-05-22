package server

import (
	"context"
	"log"
	"net"

	"topmetrics/pkg/metric"
)

const (
	maxReadBufferSize = 1024
	maxCacheSize      = 100
	maxFailedAttempts = 3
)

type Server struct {
	listener       net.Listener
	connectionType string
	ctx            context.Context
	cancel         context.CancelFunc
	metricsQueue   chan *metric.Metric
	metricHandlers map[byte]MetricHandler
	failedAttempts int
}

func NewServer(addr, connectionType string) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())

	listener, err := net.Listen(connectionType, addr)
	if err != nil {
		cancel()
		return nil, err
	}
	log.Printf("Listening on: %s", listener.Addr())
	metricsQueue := make(chan *metric.Metric)

	server := &Server{
		listener:     listener,
		ctx:          ctx,
		cancel:       cancel,
		metricsQueue: metricsQueue,
	}

	metricHandlers := map[byte]MetricHandler{
		metric.JSONType:  &LogstashMetricHandler{server},
		metric.ProtoType: &LogstashMetricHandler{server},
		metric.GOBType:   &FileMetricHandler{server},
	}
	server.metricHandlers = metricHandlers
	return server, nil
}

func (serv *Server) Run() {
	defer serv.listener.Close()

	for {
		log.Printf("Waiting connection from agent")
		conn, err := serv.listener.Accept()
		if err != nil {
			log.Printf("Error accepting: %v", err.Error())
			continue
		}
		go serv.receiveMetricHandler(conn)
	}
}

func (serv *Server) Stop() {
	log.Println("Server stopped with cancel")
	serv.cancel()
	serv.listener.Close()
	close(serv.metricsQueue)
}
