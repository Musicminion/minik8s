package serverconfig

import "miniK8s/pkg/config"

const (
	ResourceName  = "ResourceName"
	APIVersion   = "v1"
)

type ServerConfig struct {
	IfDebug  bool
	Port     int
	ListenIP string
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		IfDebug:  false,
		ListenIP: "0.0.0.0",
		Port:     config.API_Server_Port,
	}
}
