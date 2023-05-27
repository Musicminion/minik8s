package container

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"miniK8s/pkg/k8log"
	dockerclient "miniK8s/pkg/kubelet/dockerClient"
	"miniK8s/pkg/kubelet/runtime/image"
	minik8sTypes "miniK8s/pkg/minik8sTypes"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/go-connections/nat"
)

type ContainerManager struct {
}

// 创建容器的方法,返回容器的ID和错误
func (c *ContainerManager) CreateContainer(name string, option *minik8sTypes.ContainerConfig) (string, error) {
	// 获取docker的client
	k8log.DebugLog("Container Manager", "container create: "+name)
	ctx := context.Background()
	client, err := dockerclient.NewDockerClient()
	if err != nil {
		return "", err
	}
	defer client.Close()

	// 处理镜像的拉取，创建一个ImageManager
	imageManager := &image.ImageManager{}
	// 拉取镜像，根据拉取镜像的策略
	_, err = imageManager.PullImageWithPolicy(option.Image, option.ImagePullPolicy)
	if err != nil {
		return "", err
	}

	// 由于我们这些都是k8s的容器，所以我们需要给这些容器添加一些标签
	option.Labels[minik8sTypes.ContainerLabel_IfK8S] = minik8sTypes.ContainerLabel_IfK8S_True
	exposedPortSet := nat.PortSet{}
	for key := range option.ExposedPorts {
		exposedPortSet[nat.Port(key)] = struct{}{}
	}

	// 创建容器的时候需要指定容器的配置、主机配置、网络配置、存储卷配置、容器名
	result, err := client.ContainerCreate(
		ctx,
		&container.Config{
			Image:        option.Image,
			Cmd:          option.Cmd,
			Env:          option.Env,
			Tty:          option.Tty,
			Labels:       option.Labels,
			Entrypoint:   option.Entrypoint,
			Volumes:      option.Volumes,
			ExposedPorts: exposedPortSet,
		},
		&container.HostConfig{
			NetworkMode:  container.NetworkMode(option.NetworkMode),
			Binds:        option.Binds,
			PortBindings: option.PortBindings,
			IpcMode:      container.IpcMode(option.IpcMode),
			PidMode:      container.PidMode(option.PidMode),
			VolumesFrom:  option.VolumesFrom,
			Links:        option.Links,
			Resources: container.Resources{
				Memory:   option.MemoryLimit,
				NanoCPUs: option.CPUResourceLimit,
			},
		},
		nil,
		nil,
		name,
	)

	if err != nil {
		return "", err
	}

	// 将容器的ID返回
	return result.ID, nil
}

// 启动一个容器，返回容器的ID和错误
func (c *ContainerManager) StartContainer(containerID string) (string, error) {
	ctx := context.Background()
	client, err := dockerclient.NewDockerClient()
	if err != nil {
		return "", err
	}
	defer client.Close()

	err = client.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		return "", err
	}

	return containerID, nil
}

// 停止一个容器，返回容器的ID和错误
// 注意：如果多次调用这个方法，重复停止一个容器，不会报错
func (c *ContainerManager) StopContainer(containerID string) (string, error) {
	ctx := context.Background()
	client, err := dockerclient.NewDockerClient()
	if err != nil {
		return "", err
	}
	defer client.Close()

	err = client.ContainerStop(ctx, containerID, nil)
	if err != nil {
		return "", err
	}

	return containerID, nil
}

// 删除一个容器，返回容器的ID和错误
func (c *ContainerManager) RemoveContainer(containerID string) (string, error) {
	ctx := context.Background()
	client, err := dockerclient.NewDockerClient()
	if err != nil {
		return "", err
	}
	defer client.Close()

	// 检查容器是否在运行，如果在运行，先停止容器
	containerInfo, err := c.GetContainerInspectInfo(containerID)
	if err != nil {
		return "", err
	}
	if containerInfo.State != nil && containerInfo.State.Running {
		_, err = c.StopContainer(containerID)
		if err != nil {
			return "", err
		}
	}

	// 然后删除容器
	err = client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{})
	if err != nil {
		return "", err
	}

	return containerID, nil
}

// 列出所有的k8s的容器，包括已经停止的，返回容器的列表和错误
func (c *ContainerManager) ListContainers() ([]types.Container, error) {
	ctx := context.Background()
	client, err := dockerclient.NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	listFliter := filters.NewArgs()
	listFliter.Add("label", fmt.Sprint(minik8sTypes.ContainerLabel_IfK8S, "=", minik8sTypes.ContainerLabel_IfK8S_True))

	containers, err := client.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: listFliter,
	})

	if err != nil {
		return nil, err
	}

	return containers, nil
}

// 列出所有的容器，包括非k8s的容器，返回容器的列表和错误
func (c *ContainerManager) ListLocalContainers() ([]types.Container, error) {
	ctx := context.Background()
	client, err := dockerclient.NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	listFliter := filters.NewArgs()
	containers, err := client.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: listFliter,
	})

	if err != nil {
		return nil, err
	}

	return containers, nil
}

// 按照条件查询容器列表，返回容器的列表和错误
func (c *ContainerManager) ListContainersWithOpt(filter map[string][]string) ([]types.Container, error) {
	ctx := context.Background()
	client, err := dockerclient.NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	listFliter := filters.NewArgs()
	for key, valVec := range filter {
		for _, val := range valVec {
			// listFliter.Add(key, val)
			listFliter.Add("label", fmt.Sprint(key, "=", val))
		}
	}

	containers, err := client.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: listFliter,
	})

	if err != nil {
		return nil, err
	}

	return containers, nil
}

// 获取某个容器的状态，返回容器的状态和错误
// 注意：Stats反应的是容器的实时运行状态，比如CPU、内存状态，而不是什么容器是否在运行的这些
// 要获取容器是否在运行之类的状态，需要使用client.ContainerInspect
func (c *ContainerManager) GetContainerStats(containerID string) (*types.StatsJSON, error) {
	ctx := context.Background()
	client, err := dockerclient.NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	containerState, err := client.ContainerStats(ctx, containerID, false)

	if err != nil {
		return nil, err
	}
	defer containerState.Body.Close()

	decoder := json.NewDecoder(containerState.Body)
	statsInfo := &types.StatsJSON{}
	err = decoder.Decode(statsInfo)
	if err != nil {
		return nil, err
	}

	return statsInfo, nil
}

// 获取容器Inspect的信息
func (c *ContainerManager) GetContainerInspectInfo(containerID string) (*types.ContainerJSON, error) {
	ctx := context.Background()
	client, err := dockerclient.NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	containerInfo, err := client.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, err
	}

	return &containerInfo, nil
}

// 重启一个容器，返回容器的ID和错误
func (c *ContainerManager) RestartContainer(containerID string) (string, error) {
	ctx := context.Background()
	client, err := dockerclient.NewDockerClient()
	if err != nil {
		return "", err
	}
	defer client.Close()

	err = client.ContainerRestart(ctx, containerID, nil)
	if err != nil {
		return "", err
	}

	return containerID, nil
}

func (c *ContainerManager) ExecContainer(containerID string, cmd []string) (string, error) {
	k8log.DebugLog("Container Manager", "container "+containerID+"exec: "+strings.Join(cmd, " "))
	ctx := context.Background()
	client, err := dockerclient.NewDockerClient()
	if err != nil {
		return "", err
	}
	defer client.Close()

	execID, err := client.ContainerExecCreate(ctx, containerID, types.ExecConfig{
		Cmd: cmd,
	})
	if err != nil {
		return "", err
	}

	resp, err := client.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return "", err
	}
	defer resp.Close()

	var outputBuf bytes.Buffer
	_, err = io.Copy(&outputBuf, resp.Reader)
	if err != nil {
		return "", err
	}

	output := strings.TrimSpace(outputBuf.String())
	return output, nil
}

// 创建非标记为k8s的容器，返回容器的ID和错误
func (c *ContainerManager) CreateHelperContainer(name string, option *minik8sTypes.ContainerConfig) (string, error) {
	// 获取docker的client
	k8log.DebugLog("Container Manager", "container create: "+name)
	ctx := context.Background()
	client, err := dockerclient.NewDockerClient()
	if err != nil {
		return "", err
	}
	defer client.Close()

	// 处理镜像的拉取，创建一个ImageManager
	imageManager := &image.ImageManager{}
	// 拉取镜像，根据拉取镜像的策略
	_, err = imageManager.PullImageWithPolicy(option.Image, option.ImagePullPolicy)
	if err != nil {
		return "", err
	}

	// 由于我们这些都是k8s的容器，所以我们需要给这些容器添加一些标签
	// option.Labels[minik8sTypes.ContainerLabel_IfK8S] = minik8sTypes.ContainerLabel_IfK8S_True
	exposedPortSet := nat.PortSet{}
	for key := range option.ExposedPorts {
		exposedPortSet[nat.Port(key)] = struct{}{}
	}

	// 创建容器的时候需要指定容器的配置、主机配置、网络配置、存储卷配置、容器名
	result, err := client.ContainerCreate(
		ctx,
		&container.Config{
			Image:        option.Image,
			Cmd:          option.Cmd,
			Env:          option.Env,
			Tty:          option.Tty,
			Labels:       option.Labels,
			Entrypoint:   option.Entrypoint,
			Volumes:      option.Volumes,
			ExposedPorts: exposedPortSet,
		},
		&container.HostConfig{
			NetworkMode:  container.NetworkMode(option.NetworkMode),
			Binds:        option.Binds,
			PortBindings: option.PortBindings,
			IpcMode:      container.IpcMode(option.IpcMode),
			PidMode:      container.PidMode(option.PidMode),
			VolumesFrom:  option.VolumesFrom,
			Links:        option.Links,
			Resources: container.Resources{
				Memory:   option.MemoryLimit,
				NanoCPUs: option.CPUResourceLimit,
			},
		},
		nil,
		nil,
		name,
	)

	if err != nil {
		return "", err
	}

	// 将容器的ID返回
	return result.ID, nil
}

func (c *ContainerManager) CalculateContainerResource(containerID string) (float64, float64, error) {
	client, err := dockerclient.NewDockerClient()
	if err != nil {
		return 0, 0, err
	}
	defer client.Close()

	stats, err := c.GetContainerStats(containerID)
	if err != nil {
		return 0, 0, err
	}

	// 计算cpu使用率
	cpuPercent := calculateCPUPercentUnix(stats)
	// 计算memory使用率
	memoryPercent := calculateMemoryPercentUnix(stats)

	return cpuPercent, memoryPercent, nil
}

// 辅助函数，用来计算CPU使用率
func calculateCPUPercentUnix(v *types.StatsJSON) float64 {
	var (
		cpuPercent  = 0.0
		cpuDelta    = float64(v.CPUStats.CPUUsage.TotalUsage) - float64(v.PreCPUStats.CPUUsage.TotalUsage)
		systemDelta = float64(v.CPUStats.SystemUsage) - float64(v.PreCPUStats.SystemUsage)
		onlineCPUs  = float64(v.CPUStats.OnlineCPUs)
	)
	if onlineCPUs == 0.0 {
		onlineCPUs = float64(len(v.CPUStats.CPUUsage.PercpuUsage))
	}
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 1.0
	}
	return cpuPercent
}

// // 辅助函数，用来计算内存使用率
func calculateMemoryPercentUnix(v *types.StatsJSON) float64 {
	var (
		memPercent = 0.0
		memUsage   = float64(v.MemoryStats.Usage)
		memLimit   = float64(v.MemoryStats.Limit)
	)
	if memLimit > 0.0 {
		memPercent = memUsage / memLimit 
	}
	return memPercent
}

// import (
// 	"miniK8s/pkg/kubelet/containerdClient"
// 	"time"
// )

// type ContainerImagePullPolicy string

// const (
// 	ContainerImagePullPolicyAlways       ContainerImagePullPolicy = "Always"
// 	ContainerImagePullPolicyIfNotPresent ContainerImagePullPolicy = "IfNotPresent"
// 	ContainerImagePullPolicyNever        ContainerImagePullPolicy = "Never"
// 	// ContainerImagePullPolicyNever 代表从不拉取镜像，哪怕本地没有镜像
// 	// 这个时候如果本地没有镜像，而又强制创建容器，就会报错
// )

// // 定义容器的状态
// type ContainerState string

// // 在containerd中，容器的状态可以是以下值之一
// // created：容器已创建但未启动。
// // running：容器正在运行。
// // stopped：容器已停止运行。
// // paused：容器已暂停。
// // unknown：无法确定容器的状态。
// // 容器的状态可以通过使用ctr命令行工具或使用containerd的API来获取。

// const (
// 	Created ContainerState = "created"
// 	Running ContainerState = "running"
// 	Stopped ContainerState = "stopped"
// 	Paused  ContainerState = "paused"
// 	Unknown ContainerState = "unknown"
// )

// /*
// 	containerd			获取到容器的所有状态信息列表：
// 	ID：				容器的唯一标识符。
// 	PID：				容器的主进程ID。
// 	Status：			容器的状态，可以是created、running、stopped、paused、unknown中的一种。
// 	CreatedAt：			容器的创建时间。
// 	ExitStatus：		容器退出时的状态码，如果容器尚未退出，则为0。
// 	ExitTime：			容器的退出时间。
// 	Labels：			容器的标签，以键值对形式存储。
// 	Annotations：		容器的注释，以键值对形式存储。
// 	Spec：				容器的配置，包括容器的命令、参数、环境变量、资源限制等。
// 	SnapshotKey：		容器快照的键。
// 	Snapshotter：		用于创建容器快照的快照程序的名称。
// 	RootFS：			容器的根文件系统，包括类型和挂载点等信息。
// 	Image：				用于创建容器的镜像信息，包括名称、标签、摘要、大小等。
// 	Processes：			容器内所有进程的列表，包括每个进程的ID、状态、启动时间等信息。
// 	Stats：				容器的资源使用统计信息，包括CPU使用情况、内存使用情况、网络使用情况等。
// 	Platform：			容器运行的平台，包括操作系统和硬件架构等信息。
// 	PortMappings：		容器的端口映射列表，包括每个端口映射的协议、主机端口、容器端口等信息
// */

// // 定义一个容器的结构体
// type Container struct {
// 	// 容器的唯一标识符
// 	ID string
// 	// 容器的主进程ID
// 	PID int
// 	// 容器的状态
// 	Status ContainerState
// 	// 容器的创建时间
// 	CreatedAt time.Time
// 	// 容器退出时的状态码，如果容器尚未退出，则为0
// 	ExitStatus int
// 	// 容器的退出时间
// 	ExitTime time.Time
// 	// 容器的标签，以键值对形式存储
// 	Labels map[string]string
// 	// 容器的注释，以键值对形式存储
// 	Annotations map[string]string
// }

// func (c *Container) CreateContainer(name string) (string, error) {
// 	client, err := containerdClient.NewContainerdClient()

// 	if err != nil {
// 		return "", err
// 	}

// 	defer client.Close()

// 	// 创建容器
// 	// container, err := client.NewContainer(context.Background(), name,
// 	// 	containerd.WithImage("docker.io/library/busybox:latest"),
// 	// )

// 	return "", nil
// }
