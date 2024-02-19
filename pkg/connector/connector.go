package connector

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"
)

const sendTimeout = 3 * time.Second

type Sender interface {
	SendMetric(ctx context.Context, data []byte) error
	initConnection(ctx context.Context) error
}

type Connector struct {
	Host, Port, connectionType string
	net.Conn
}

func NewConnector(host string, port string, connectionType string) *Connector {
	return &Connector{Host: host, Port: port, connectionType: connectionType}
}

func (c *Connector) Connect(ctx context.Context) (net.Conn, error) {
	host := net.JoinHostPort(c.Host, c.Port)
	log.Println("Try to connect to", host)
	dialer := &net.Dialer{Timeout: sendTimeout}
	conn, err := dialer.DialContext(ctx, c.connectionType, host)
	if err != nil {
		return nil, fmt.Errorf("connection error: failed to connect to server at %s: %w", host, err)
	}
	log.Printf("Connected to: %s", conn.RemoteAddr())
	return conn, nil
}

func (c *Connector) initConnection(ctx context.Context) error {
	config := NewConnector(c.Host, c.Port, c.connectionType)
	conn, err := config.Connect(ctx)
	if err != nil {
		return err
	}
	c.Conn = conn
	return nil
}

func (c *Connector) SendMetric(ctx context.Context, data []byte) error {
	if c.Conn == nil {
		if err := c.initConnection(ctx); err != nil {
			return fmt.Errorf("initialize error: %w", err)
		}
	}
	numBytes, err := c.Write(data)
	if err != nil {
		c.Close()
		c.Conn = nil
		return fmt.Errorf("send error: failed to send log message: %w", err)
	}
	log.Printf("Sent: %d bytes\n", numBytes)
	return nil
}
