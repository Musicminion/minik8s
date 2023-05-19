// 把etcd要存储的目标目录配置卸载这里

package serverconfig

import "time"

const (
	EtcdTokenPath = "/registry/tokens/"

	// 完整路径：/registry/nodes/<node-name>
	EtcdNodePath = "/registry/nodes/"

	// 完整路径：/registry/pods/<namespace>/<pod-name>
	EtcdPodPath = "/registry/pods/"

	// 完整路径：/registry/services/<svc-name>
	EtcdServicePath = "/registry/services/"

	// 完整路径：/registry/svclabel/<label-key>/<label-value>/<svc-uuid>
	// 完整路径：/registry/svclabels/<label-key>/<label-value>/<svc-uuid>
	EtcdServiceSelectorPath = "/registry/svclabels/"

	// 完整路径：/registry/endpoints/<label-key>/<label-value>/<endpoint-uuid> ?
	// 完整路径：/registry/endpoints/<label-key>/<label-value>/<pod-uuid> ?
	EndpointPath = "/registry/endpoints/"

	// 完整路径：/registry/jobs/<namespace>/<job-name>
	EtcdJobPath = "/registry/jobs/"

	// 完整路径：/registry/jobfile/<namespace>/<job-name>
	EtcdJobFilePath = "/registry/jobfile/"
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
