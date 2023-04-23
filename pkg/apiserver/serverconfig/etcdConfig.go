// 把etcd要存储的目标目录配置卸载这里

package serverconfig

import "time"

const (
	EtcdTokenPath   = "/registry/tokens/"
	EtcdNodePath    = "/registry/nodes/"
	EtcdPodPath     = "/registry/pods/"
	EtcdServicePath = "/registry/services/"
)

type EtcdConfig struct {
	EtcdEndpoints []string
	EtcdTimeout   time.Duration
}

func DefaultEtcdConfig() *EtcdConfig {
	return &EtcdConfig{
		EtcdEndpoints: []string{"localhost:2379"},
		EtcdTimeout:   5 * time.Second,
	}
}
