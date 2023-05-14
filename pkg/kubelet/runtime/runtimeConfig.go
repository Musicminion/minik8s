package runtime

import "miniK8s/pkg/apiObject"

const (
	// PauseContainerName pause容器的名字基础-后面会加上其他信息
	PauseContainerNameBase = "pause-"
	// PauseContainerImage pause容器的镜像
	PauseContainerImage = "registry.aliyuncs.com/google_containers/pause:3.6"
	// Regular容器的名字基础-后面会加上其他信息
	RegularContainerNameBase = "regular-"

	// 引用容器的时候，加上这个前缀，表示引用的是容器
	// 比如你要引用容器的ID，就是container:xxxx
	ContianerREfPrefix = "container:"

	//
)

// 用作给GetRuntimeAllPodStatus函数作为返回，返回的时候包含Pod的ID、Pod的名字、Pod的命名空间、Pod的状态
// 这样才能方便上层的调用者能够知道是哪个Pod，然后发送给对应的URL请求，更新对应的Pod的状态
type RunTimePodStatus struct {
	// Pod的ID
	PodID string
	// Pod的名字
	PodName string
	// Pod的命名空间
	PodNamespace string
	// Pod的状态
	PodStatus apiObject.PodStatus
}
