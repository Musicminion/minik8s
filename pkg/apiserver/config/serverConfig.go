package config

import (
	"time"
)

const (
	ResourceName = "ResourceName"
)

type ServerConfig struct {
	Port          int
	EtcdEndpoints []string
	EtcdTimeout   time.Duration
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:          8080,
		EtcdEndpoints: []string{"localhost:2379"},
		EtcdTimeout:   5 * time.Second,
	}
}
