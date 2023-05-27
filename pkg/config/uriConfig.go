package config

import "miniK8s/pkg/apiObject"

// 考虑到APIServer用URL，而Kuble用URI，那URI的规定就该放在全局配置里面

const (
	// URI_PROTOCOL_SCHEME
	URI_PROTOCOL_SCHEME = "http://"
)

// 这里是包括了API Server的所有URL

// Node相关操作的URL
// 根据K8s官方文档，Node是属于集群级别的资源，所以不需要namespace！
// https://kubernetes.io/zh-cn/docs/concepts/overview/working-with-objects/namespaces/#not-all-objects-are-in-a-namespace
const (
	// 请把所有和namespace【没有关系】的放在下面
	// Node是属于集群级别的资源，需要放在下面，没有名字空间
	// 所有Node状态
	NodesURL = "/api/v1/nodes"
	// 某个特定的Node状态
	NodeSpecURL = "/api/v1/nodes/:name"
	// 某个特定的Node的status
	NodeSpecStatusURL = "/api/v1/nodes/:name/status"
	// 某个特定的Node持有的所有的Pod
	NodeAllPodsURL = "/api/v1/nodes/:name/pods"

	// 请把所有和名字空间【有关系】的放在下面
	// Pod相关操作的URL
	// 获取全局的Pod的URL
	GlobalPodsURL = "/api/v1/pods"
	// 所有Pod的状态的URL
	PodsURL = "/api/v1/namespaces/:namespace/pods"
	// 某个特定Pod的URL
	PodSpecURL = "/api/v1/namespaces/:namespace/pods/:name"
	// 获取Pod的某个状态的URL
	PodSpecStatusURL = "/api/v1/namespaces/:namespace/pods/:name/status"

	// Service相关操作的URL
	// 所有Service的状态的URL
	ServiceURL = "/api/v1/namespaces/:namespace/services"
	// 某个特定Service的URL
	ServiceSpecURL = "/api/v1/namespaces/:namespace/services/:name"
	// 获取Service的某个状态的URL
	ServiceSpecStatusURL = "/api/v1/namespaces/:namespace/services/:name/status"

	// Endpoint相关操作的URL
	// 所有Endpoint的状态的URL
	EndpointURL = "/api/v1/namespaces/:namespace/endpoints"
	// 某个特定Endpoint的URL
	EndpointSpecURL = "/api/v1/namespaces/:namespace/services/:name"

	// Job相关操作的URL
	JobsURL = "/apis/v1/namespaces/:namespace/jobs"
	// 某个特定Job的URL
	JobSpecURL = "/apis/v1/namespaces/:namespace/jobs/:name"
	// 获取Job的某个状态的URL
	JobSpecStatusURL = "/apis/v1/namespaces/:namespace/jobs/:name/status"
	// Job的文件的URL
	JobFileURL = "/apis/v1/namespaces/:namespace/jobfiles"
	// 某个特定Job的文件的URL
	JobFileSpecURL = "/apis/v1/namespaces/:namespace/jobfiles/:name"

	//获取所有的 ReplicaSet，仅供controller使用，他会抓取所有的ReplicaSet，然后进行比对
	GlobalReplicaSetsURL = "/apis/v1/replicasets"
	// ReplicaSet相关操作的URL
	ReplicaSetsURL = "/apis/v1/namespaces/:namespace/replicasets"
	// 某个特定ReplicaSet的URL
	ReplicaSetSpecURL = "/apis/v1/namespaces/:namespace/replicasets/:name"
	// 获取ReplicaSet的某个状态的URL
	ReplicaSetSpecStatusURL = "/apis/v1/namespaces/:namespace/replicasets/:name/status"

	// Dns相关的URL
	// 所有Dns的URL
	DnsURL = "/apis/v1/namespaces/:namespace/dns"
	// 某个特定Dns的URL
	DnsSpecURL = "/apis/v1/namespaces/:namespace/dns/:name"

	// HPA相关的URL
	// 全局HPA的URL
	GlobalHPAURL = "/apis/v1/hpa"
	// 所有HPA的URL
	HPAURL = "/apis/v1/namespaces/:namespace/hpa"
	// 某个特定HPA的URL
	HPASpecURL = "/apis/v1/namespaces/:namespace/hpa/:name"
	// 获取HPA的某个状态的URL
	HPASpecStatusURL = "/apis/v1/namespaces/:namespace/hpa/:name/status"

	// Function相关的URL
	// 全局Function的URL
	GlobalFunctionsURL = "/apis/v1/functions"
	// 所有Function的URL(Namespace级别)
	FunctionURL = "/apis/v1/namespaces/:namespace/functions"
	// 某个特定Function的URL
	FunctionSpecURL = "/apis/v1/namespaces/:namespace/functions/:name"

	// Workflow相关的URL
	// 全局Workflow的URL
	GlobalWorkflowsURL = "/apis/v1/workflows"
	// 所有Workflow的URL(Namespace级别)
	WorkflowURL = "/apis/v1/namespaces/:namespace/workflows"
	// 某个特定Workflow的URL
	WorkflowSpecURL = "/apis/v1/namespaces/:namespace/workflows/:name"
	// Workflow的Status的URL
	WorkflowSpecStatusURL = "/apis/v1/namespaces/:namespace/workflows/:name/status"
)

const (
	// 请把所有【参数】相关的放在下面，这部分是不带冒号的
	URL_PARAM_NAME      = "name"
	URL_PARAM_NAMESPACE = "namespace"

	// 请把所有【参数】相关的放在下面，【PART】是指URI里面带冒号的部分
	URL_PARAM_NAME_PART      = ":name"
	URL_PARAM_NAMESPACE_PART = ":namespace"
)

// kind->返回所有资源的URL(给定namespace)
var ApiResourceMap = map[string]string{
	apiObject.PodKind:        PodsURL,
	apiObject.ServiceKind:    ServiceURL,
	apiObject.DnsKind:        DnsURL,
	apiObject.NodeKind:       NodesURL,
	apiObject.JobKind:        JobsURL,
	apiObject.ReplicaSetKind: ReplicaSetsURL,
	apiObject.HpaKind:        HPAURL,
	apiObject.FunctionKind:   FunctionURL,
}

// kind->返回特定资源的URL(给定namespace)
var ApiSpecResourceMap = map[string]string{
	apiObject.PodKind:        PodSpecURL,
	apiObject.ServiceKind:    ServiceSpecURL,
	apiObject.DnsKind:        DnsSpecURL,
	apiObject.NodeKind:       NodeSpecURL,
	apiObject.JobKind:        JobSpecURL,
	apiObject.ReplicaSetKind: ReplicaSetSpecURL,
	apiObject.HpaKind:        HPASpecURL,
	apiObject.FunctionKind:   FunctionSpecURL,
}
