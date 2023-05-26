package serverconfig

const (
	ResourceName  = "ResourceName"
	RequestPrefix = "http://0.0.0.0:8090"
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
		Port:     8090,
	}
}
