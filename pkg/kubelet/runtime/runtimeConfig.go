package runtime

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
