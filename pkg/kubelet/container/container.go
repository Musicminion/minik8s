package container

import (
	"miniK8s/pkg/kubelet/containerdClient"
	"time"
)

type ContainerImagePullPolicy string

const (
	ContainerImagePullPolicyAlways       ContainerImagePullPolicy = "Always"
	ContainerImagePullPolicyIfNotPresent ContainerImagePullPolicy = "IfNotPresent"
	ContainerImagePullPolicyNever        ContainerImagePullPolicy = "Never"
	// ContainerImagePullPolicyNever 代表从不拉取镜像，哪怕本地没有镜像
	// 这个时候如果本地没有镜像，而又强制创建容器，就会报错
)

// 定义容器的状态
type ContainerState string

// 在containerd中，容器的状态可以是以下值之一
// created：容器已创建但未启动。
// running：容器正在运行。
// stopped：容器已停止运行。
// paused：容器已暂停。
// unknown：无法确定容器的状态。
// 容器的状态可以通过使用ctr命令行工具或使用containerd的API来获取。

const (
	Created ContainerState = "created"
	Running ContainerState = "running"
	Stopped ContainerState = "stopped"
	Paused  ContainerState = "paused"
	Unknown ContainerState = "unknown"
)

/*
	containerd			获取到容器的所有状态信息列表：
	ID：				容器的唯一标识符。
	PID：				容器的主进程ID。
	Status：			容器的状态，可以是created、running、stopped、paused、unknown中的一种。
	CreatedAt：			容器的创建时间。
	ExitStatus：		容器退出时的状态码，如果容器尚未退出，则为0。
	ExitTime：			容器的退出时间。
	Labels：			容器的标签，以键值对形式存储。
	Annotations：		容器的注释，以键值对形式存储。
	Spec：				容器的配置，包括容器的命令、参数、环境变量、资源限制等。
	SnapshotKey：		容器快照的键。
	Snapshotter：		用于创建容器快照的快照程序的名称。
	RootFS：			容器的根文件系统，包括类型和挂载点等信息。
	Image：				用于创建容器的镜像信息，包括名称、标签、摘要、大小等。
	Processes：			容器内所有进程的列表，包括每个进程的ID、状态、启动时间等信息。
	Stats：				容器的资源使用统计信息，包括CPU使用情况、内存使用情况、网络使用情况等。
	Platform：			容器运行的平台，包括操作系统和硬件架构等信息。
	PortMappings：		容器的端口映射列表，包括每个端口映射的协议、主机端口、容器端口等信息
*/

// 定义一个容器的结构体
type Container struct {
	// 容器的唯一标识符
	ID string
	// 容器的主进程ID
	PID int
	// 容器的状态
	Status ContainerState
	// 容器的创建时间
	CreatedAt time.Time
	// 容器退出时的状态码，如果容器尚未退出，则为0
	ExitStatus int
	// 容器的退出时间
	ExitTime time.Time
	// 容器的标签，以键值对形式存储
	Labels map[string]string
	// 容器的注释，以键值对形式存储
	Annotations map[string]string
}

func (c *Container) CreateContainer(name string) (string, error) {
	client, err := containerdClient.NewContainerdClient()

	if err != nil {
		return "", err
	}

	defer client.Close()

	// 创建容器
	// container, err := client.NewContainer(context.Background(), name,
	// 	containerd.WithImage("docker.io/library/busybox:latest"),
	// )

	return "", nil
}
