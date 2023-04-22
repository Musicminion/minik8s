package config

// 这里是包括了API Server的所有URL

// Node相关操作的URL
// 根据K8s官方文档，Node是属于集群级别的资源，所以不需要namespace！
// https://kubernetes.io/zh-cn/docs/concepts/overview/working-with-objects/namespaces/#not-all-objects-are-in-a-namespace
const (
	NodeURL                  = "/api/v1/nodes/"
	NodeURLWithSpecifiedName = "/api/v1/nodes/:name"
)
