package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"topmetrics/pkg/connector"
	"topmetrics/pkg/metric"
)

const (
	// Default connection settings
	defaultHost           = "192.168.0.199"
	defaultPort           = "8080"
	defaultConnectionType = "tcp"
	maxFailedAttempts     = 3
	reconnectTimeout      = maxFailedAttempts * 500 * time.Millisecond

	// Default agent settings
	defaultInterval      = 2 * time.Second
	defaultTimeout       = 10 * time.Minute
	defaultCount         = 5
	defaultSerializeType = "proto"

	// Config path
	cfgPath = "configs/agent/default_cfg_proto.json"
)

type Agent struct {
	*connector.Connector
	*Config
	serializeTypes map[string]byte
	Ctx            context.Context
	Cancel         context.CancelFunc
}

type Config struct {
	Host          string        `json:"host,omitempty"`
	Port          string        `json:"port,omitempty"`
	Protocol      string        `json:"protocol,omitempty"`
	Hostname      string        `json:"hostname,omitempty"`
	SerializeType string        `json:"type,omitempty"`
	Interval      time.Duration `json:"interval,omitempty"`
	Timeout       time.Duration `json:"timeout,omitempty"`
	Count         int           `json:"count,omitempty"`
	FailAttempts  int           `json:",omitempty"`
}

func NewConfig() (*Agent, error) {
	agnt := &Agent{}
	agnt.serializeTypes = map[string]byte{
		"j":     metric.JSONType,
		"json":  metric.JSONType,
		"g":     metric.GOBType,
		"gob":   metric.GOBType,
		"p":     metric.ProtoType,
		"proto": metric.ProtoType,
	}

	cfg, err := agnt.getConfigFile()
	if err != nil {
		return nil, err
	}
	agnt.Config = cfg

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	agnt.Ctx = ctx
	agnt.Cancel = cancel

	conn := connector.NewConnector(cfg.Host, cfg.Port, cfg.Protocol)
	agnt.Connector = conn

	if err = agnt.SetHostname(cfg.Hostname); err != nil {
		return nil, err
	}
	return agnt, nil
}

func (agnt *Agent) SetHostname(hostname string) error {
	if hostname != "" {
		agnt.Config.Hostname = hostname
		return nil
	}
	name, err := os.Hostname()
	if err != nil {
		return err
	}
	agnt.Config.Hostname = name
	return nil
}

func (agnt *Agent) getDataType() byte {
	dataType, ok := agnt.serializeTypes[strings.ToLower(agnt.SerializeType)]
	if !ok {
		log.Println("Use default serializer: Proto")
		return metric.ProtoType
	}
	return dataType
}

func (agnt *Agent) getConfigFile() (*Config, error) {
	cfg := &Config{}
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		if err = cfg.createFile(); err != nil {
			return nil, fmt.Errorf("config error: create file error: %w", err)
		}
		return cfg, nil
	}
	conf, err := os.Open(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}
	defer func(conf *os.File) {
		err := conf.Close()
		if err != nil {
			return
		}
	}(conf)
	if err = json.NewDecoder(conf).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("config decode error: %w", err)
	}
	return cfg, nil
}

func (cfg *Config) createFile() error {
	const fileName = "./config.json"
	cfg.setDefaults()
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("config encode error: %w", err)
	}
	if err = os.WriteFile(fileName, data, 0600); err != nil {
		return fmt.Errorf("config write error: %w", err)
	}
	log.Printf("The config file is not exist. Use default values from %s\n", fileName)
	return nil
}

func (cfg *Config) UnmarshalJSON(data []byte) error {
	type Alias Config
	aux := &struct {
		Interval string `json:"interval,omitempty"`
		Timeout  string `json:"timeout,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(cfg),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	interval, err := time.ParseDuration(aux.Interval)
	if err != nil {
		return err
	}
	cfg.Interval = interval
	timeout, err := time.ParseDuration(aux.Timeout)
	if err != nil {
		return err
	}
	cfg.Timeout = timeout
	return nil
}

func (cfg *Config) setDefaults() {
	if cfg.Host == "" {
		cfg.Host = defaultHost
	}
	if cfg.Port == "" {
		cfg.Port = defaultPort
	}
	if cfg.Protocol == "" {
		cfg.Protocol = defaultConnectionType
	}
	if cfg.SerializeType == "" {
		cfg.SerializeType = defaultSerializeType
	}
	if cfg.Interval == 0 {
		cfg.Interval = defaultInterval
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.Count == 0 {
		cfg.Count = defaultCount
	}
}
