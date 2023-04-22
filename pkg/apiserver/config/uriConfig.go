package config

// 这里是包括了API Server的所有URL

// Node相关操作的URL
// 根据K8s官方文档，Node是属于集群级别的资源，所以不需要namespace！
// https://kubernetes.io/zh-cn/docs/concepts/overview/working-with-objects/namespaces/#not-all-objects-are-in-a-namespace
const (
	// 所有Node状态
	NodesURL = "/api/v1/nodes/"
	// 某个特定的Node状态
	NodeSpecURL = "/api/v1/nodes/:name"
	// 某个特定的Node的status
	NodeSpecStatusURL = "/api/v1/nodes/:name/status"
)
