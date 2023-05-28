package dockerregistry

import (
	"miniK8s/pkg/config"

	"github.com/docker/docker/api/types"
)

const (
	// RegistryImageName 镜像管理中心的镜像名称
	Registry_Image_Name = "registry:2.8.2"

	Registry_Server_BindIP         = "0.0.0.0"
	Registry_Server_Port           = "5000"
	Registry_Server_Port_Protocol  = "5000/tcp"
	Registry_Server_Container_Name = "minik8s-registry"
	Registry_Server_Username       = "example"
	Registry_Server_Password       = "example"
)

var Registry_Server_IP = config.GetMasterIP()
var Registry_Server_Prefix = Registry_Server_IP + ":" + Registry_Server_Port

var (
	authInfo = types.AuthConfig{
		Username: Registry_Server_Username,
		Password: Registry_Server_Password,
	}
)
