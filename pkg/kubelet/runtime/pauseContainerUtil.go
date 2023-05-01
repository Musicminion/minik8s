package runtime

import (
	"fmt"
	"miniK8s/pkg/apiObject"
	minik8sTypes "miniK8s/pkg/minik8sTypes"

	"miniK8s/util/netutil"
	"miniK8s/util/uuid"
	"miniK8s/util/weave"

	"github.com/docker/go-connections/nat"
)

// ************************************************************

func (r *runtimeManager) removePauseContainer(pod *apiObject.PodStore) (string, error) {
	// TODO: 从容器管理器中删除pause容器
	filter := make(map[string][]string)
	// filter[minik8sTypes.ContainerLabel_Pod] = []string{pod.Metadata.Name}
	// 在filter中添加标签
	filter[minik8sTypes.ContainerLabel_Pod] = []string{pod.Metadata.Name}

	res, err := r.containerManager.ListContainersWithOpt(filter)

	if err != nil {
		return "", err
	}

	retID := ""

	// 遍历删除pause容器
	for _, container := range res {
		retID = container.ID
		// 删除pause容器
		if _, err := r.containerManager.RemoveContainer(container.ID); err != nil {
			return "", err
		}
	}

	return retID, nil

	// 本来打算严格要求是1，现在打算不严格要求，哪怕不存在也会正常返回
	// if len(res) != 1 {
	// 	return "", fmt.Errorf("pause container found more than one")
	// }

	// // 删除pause容器
	// if _, err := r.containerManager.RemoveContainer(res[0].ID); err != nil {
	// 	return "", err
	// }

	// return res[0].ID, nil
}

// 获取pause容器的创建配置信息
func (r *runtimeManager) getPauseContainerConfig(pod *apiObject.PodStore) (*minik8sTypes.ContainerConfig, error) {
	// [获取暴露端口] 根据Pod的配置文件，获取pause容器的暴露端口
	PodAllPortsBinds := nat.PortMap{}
	PodAllPortsSet := map[string]struct{}{}

	// [容器内部的端口] 放行容器内部的端口
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			// 检查HOSTIP、HOSTPORT、PROTOCOL是否为空
			if port.HostIP == "" {
				port.HostIP = minik8sTypes.PORT_LOCALHOST_IP
			}
			if port.Protocol == "" {
				port.Protocol = minik8sTypes.PORT_PROTOCOL_TCP
			}
			if port.HostPort == "" {
				avaiPort, err := netutil.GetAvailablePort()
				if err != nil {
					return nil, err
				}
				port.HostPort = avaiPort
			}

			// 组装一个port和协议的字符串
			portBindingKey, err := nat.NewPort(port.Protocol, port.ContainerPort)
			if err != nil {
				return nil, err
			}

			// 检查PodAllPortsBinds是否存在该端口，存在就说明出现了端口冲突
			if _, ok := PodAllPortsBinds[portBindingKey]; ok {
				return nil, fmt.Errorf("port conflict")
			}

			// 绑定端口
			PodAllPortsBinds[portBindingKey] = []nat.PortBinding{
				{
					HostIP:   port.HostIP,
					HostPort: port.HostPort,
				},
			}

			// 本来打算采取下面的写法，但是发现这样写可能会导致所有的容器都绑定到同一个端口上
			// PodAllPortsBinds[string(portBindingKey)].append(nat.PortBinding{
			// 	HostIP:   port.HostIP,
			// 	HostPort: port.HostPort,
			// })
		}
	}

	// [容器标签] 为pause容器添加标签
	// 遍历pod的标签，将pod的标签添加到pause容器的标签中
	pauseLabels := map[string]string{}
	for labelKey, labelVal := range pod.Metadata.Labels {
		pauseLabels[labelKey] = labelVal
	}
	// 为pause容器添加标签，标签的key为"pod"，value为pod的名字
	pauseLabels[minik8sTypes.ContainerLabel_Pod] = pod.Metadata.Name

	config := minik8sTypes.ContainerConfig{
		Image:        PauseContainerImage,
		Labels:       pauseLabels,
		PortBindings: PodAllPortsBinds,
		ExposedPorts: PodAllPortsSet,
		Volumes:      nil,
		Env:          nil,
		IpcMode:      minik8sTypes.Contianer_IPCMode_Sharable,
	}

	return &config, nil
}

func (r *runtimeManager) createPauseContainer(pod *apiObject.PodStore) (string, error) {
	// [镜像检查] 检查pause镜像是否存在，不存在则拉取
	_, err := r.imageManager.PullImageWithPolicy(PauseContainerImage, minik8sTypes.PullIfNotPresent)
	if err != nil {
		return "", err
	}

	// [获取配置] 获取pause容器的创建配置信息
	pauseConfig, err := r.getPauseContainerConfig(pod)

	if err != nil {
		return "", err
	}

	// [容器创建] 创建pause容器
	// 产生一个随机的uuid
	uuid := uuid.NewUUID()

	// 取uuid的前12位作为pause容器的名字
	newPauseName := fmt.Sprintf("%s-%s", PauseContainerNameBase, uuid)

	ID, err := r.containerManager.CreateContainer(newPauseName, pauseConfig)

	if err != nil {
		return "", err
	}

	// [容器启动] 启动pause容器
	_, err = r.containerManager.StartContainer(ID)

	if err != nil {
		return "", err
	}

	// [Weave网络] 为pause容器添加网络
	weave.WeaveAttach(ID, pod.Status.PodIP)

	return ID, nil
}

// 启动一个
// func (r *runtimeManager) startPauseContainer(pod *apiObject.PodStore) (string, error) {
// 	var filter = make(map[string][]string)
// 	// filter[minik8sTypes.ContainerLabel_Pod] = []string{pod.Metadata.Name}
// 	// 在filter中添加标签
// 	filter[minik8sTypes.ContainerLabel_Pod] = []string{pod.Metadata.Name}

// 	res, err := r.containerManager.ListContainersWithOpt(filter)

// 	if err != nil {
// 		return "", err
// 	}

// 	if len(res) != 1 {
// 		return "", fmt.Errorf("pause container found more than one")
// 	}

// 	// 启动pause容器
// 	if _, err := r.containerManager.StartContainer(res[0].ID); err != nil {
// 		return "", err
// 	}

// 	return res[0].ID, nil
// }
