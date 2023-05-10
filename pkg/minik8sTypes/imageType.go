package minik8stypes

// ImagePullPolicy表示拉取镜像的策略
type ImagePullPolicy string

// 三个可能的值
const (
	PullIfNotPresent ImagePullPolicy = "IfNotPresent"
	PullAlways       ImagePullPolicy = "Always"
	PullNever        ImagePullPolicy = "Never"
)
