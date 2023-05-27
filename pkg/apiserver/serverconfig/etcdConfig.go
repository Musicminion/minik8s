// 把etcd要存储的目标目录配置卸载这里

package serverconfig

import "time"

const (
	//用来记录分配的最大的ServiceIP
	EtcdIPPath = "/registry/allocatedIP"

	EtcdTokenPath = "/registry/tokens/"

	// 完整路径：/registry/nodes/<node-name>
	EtcdNodePath = "/registry/nodes/"

	// 完整路径：/registry/pods/<namespace>/<pod-name>
	EtcdPodPath = "/registry/pods/"
	// 完整路径：/registry/services/<namespace>/<svc-name>
	EtcdServicePath = "/registry/services/"

	// 完整路径：/registry/svclabels/<label-key>/<label-value>/<svc-uuid>
	EtcdServiceSelectorPath = "/registry/svclabels/"

	// 完整路径：/registry/endpoints/<label-key>/<label-value>/<pod-uuid> ?
	EndpointPath = "/registry/endpoints/"

	// 完整路径：/registry/jobs/<namespace>/<job-name>
	EtcdJobPath = "/registry/jobs/"

	// 完整路径：/registry/jobfile/<namespace>/<job-name>
	EtcdJobFilePath = "/registry/jobfile/"

	// 完整路径：/registry/replicasets/<namespace>/<replicaset-name>
	EtcdReplicaSetPath = "/registry/replicasets/"

	// 完整路径：/registry/dns/<namespace>/<dns-name>
	EtcdDnsPath = "/registry/dns/"

	// 完整路径：/registry/hpa/<namespace>/<hpa-name>
	EtcdHpaPath = "/registry/hpa/"

	// 完整路径：/registry/function/<namespace>/<function-name>
	EtcdFunctionPath = "/registry/function/"

	// 完整路径：/registry/workflows/<namespace>/<workflow-name>
	EtcdWorkflowPath = "/registry/workflows/"
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
