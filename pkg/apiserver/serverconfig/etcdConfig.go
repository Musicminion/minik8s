// 把etcd要存储的目标目录配置卸载这里

package serverconfig

import "time"

const (
	EtcdTokenPath = "/registry/tokens/"
	EtcdNodePath  = "/registry/nodes/"

	// 完整路径：/registry/pods/<namespace>/<pod-name>
	EtcdPodPath = "/registry/pods/"
	// 完整路径：/registry/services/<svc-name>
	EtcdServicePath = "/registry/services/"
	// 完整路径：/registry/svclabels/<label-key>/<label-value>/<svc-uuid>
	EtcdServiceSelectorPath = "/registry/svclabels/"
	// 完整路径：/registry/endpoints/<label-key>/<label-value>/<pod-uuid>
	EndpointPath = "/registry/endpoints/"
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
