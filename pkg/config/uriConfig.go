package config

// 考虑到APIServer用URL，而Kuble用URI，那URI的规定就该放在全局配置里面

// 这里是包括了API Server的所有URL

// Node相关操作的URL
// 根据K8s官方文档，Node是属于集群级别的资源，所以不需要namespace！
// https://kubernetes.io/zh-cn/docs/concepts/overview/working-with-objects/namespaces/#not-all-objects-are-in-a-namespace
const (
	// 请把所有根名字空间【没有关系】的放在下面
	// Node是属于集群级别的资源，需要放在下面，没有名字空间
	// 所有Node状态
	NodesURL = "/api/v1/nodes/"
	// 某个特定的Node状态
	NodeSpecURL = "/api/v1/nodes/:name"
	// 某个特定的Node的status
	NodeSpecStatusURL = "/api/v1/nodes/:name/status"

	// 请把所有根名字空间【有关系】的放在下面
	// Pod相关操作的URL
	// 所有Pod的状态的URL
	PodsURL = "/api/v1/namespaces/:namespace/pods"
	// 某个特定Pod的URL
	PodSpecURL = "/api/v1/namespaces/:namespace/pods/:name"
	// 获取Pod的某个状态的URL
	PodSpecStatusURL = "/api/v1/namespaces/:namespace/pods/:name/status"
)