package config

const (
	ResourceName = "ResourceName"
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
		Port:     8090,
	}
}
