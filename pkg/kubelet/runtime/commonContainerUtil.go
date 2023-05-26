package runtime

import (
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/k8log"
	minik8sTypes "miniK8s/pkg/minik8sTypes"

	"miniK8s/util/uuid"
)

// **********************************************************************************
//  一组容器的操作放在下面
// **********************************************************************************

// 给一个Pod创建所有的Contianer
func (r *runtimeManager) createPodAllContainer(pod *apiObject.PodStore, pauseContainerID string) (string, error) {
	// TODO: 从容器管理器中启动一个pod中的所有容器
	for _, container := range pod.Spec.Containers {
		_, err := r.createPodContainer(pod, &container, pauseContainerID)
		if err != nil {
			// 可能需要注意垃圾回收机制
			r.removePodAllContainer(pod)
			return "", err
		}
	}

	k8log.InfoLog("kubelet", fmt.Sprintf("create pod %s/%s success", pod.Metadata.Namespace, pod.Metadata.Name))
	return pod.Metadata.UUID, nil
}

// 会尝试删除所有的容器，如果遇到某个容器删除失败，则会继续删除其他容器，最后返回错误信息
func (r *runtimeManager) removePodAllContainer(pod *apiObject.PodStore) (string, error) {
	// TODO: 从容器管理器中删除一个pod中的所有容器
	retErr := ""

	for _, container := range pod.Spec.Containers {
		_, err := r.removePodContainer(pod, &container)
		if err != nil {
			// return "", err
			retErr += err.Error() + "\n"
			k8log.ErrorLog("kubelet", fmt.Sprintf("remove pod %s/%s failed, err %s", pod.Metadata.Namespace, pod.Metadata.Name, err.Error()))
		}
	}

	if retErr != "" {
		return "", fmt.Errorf(retErr)
	}

	k8log.InfoLog("kubelet", fmt.Sprintf("remove pod %s/%s success", pod.Metadata.Namespace, pod.Metadata.Name))
	return pod.Metadata.UUID, nil
}

// 会尝试停止所有的容器，如果遇到某个容器停止失败，则会继续停止其他容器，最后返回错误信息
func (r *runtimeManager) stopPodAllContainer(pod *apiObject.PodStore) (string, error) {
	// TODO: 从容器管理器中删除一个pod中的所有容器
	retErr := ""
	for _, container := range pod.Spec.Containers {
		_, err := r.stopPodContainer(pod, &container)
		if err != nil {
			retErr += err.Error() + "\n"
			k8log.ErrorLog("kubelet", fmt.Sprintf("stop pod %s/%s failed, err %s", pod.Metadata.Namespace, pod.Metadata.Name, err.Error()))
		}
	}

	if retErr != "" {
		return "", fmt.Errorf(retErr)
	}

	k8log.InfoLog("kubelet", fmt.Sprintf("stop pod %s/%s success", pod.Metadata.Namespace, pod.Metadata.Name))
	return pod.Metadata.UUID, nil
}

// 会尝试启动所有的容器，如果遇到某个容器启动失败，则会继续启动其他容器，最后返回错误信息
func (r *runtimeManager) startPodAllContainer(pod *apiObject.PodStore) (string, error) {
	// TODO: 从容器管理器中启动一个pod中的所有容器
	retErr := ""
	for _, container := range pod.Spec.Containers {
		_, err := r.startPodContainer(pod, &container)
		if err != nil {
			retErr += err.Error() + "\n"
			k8log.ErrorLog("kubelet", fmt.Sprintf("start pod %s/%s failed, err %s", pod.Metadata.Namespace, pod.Metadata.Name, err.Error()))
		}
	}

	if retErr != "" {
		return "", fmt.Errorf(retErr)
	}

	k8log.InfoLog("kubelet", fmt.Sprintf("start pod %s/%s success", pod.Metadata.Namespace, pod.Metadata.Name))
	return pod.Metadata.UUID, nil
}

// 尝试重启所有的容器，如果遇到某个容器重启失败，则会继续重启其他容器，最后返回错误信息
func (r *runtimeManager) restartPodAllContainer(pod *apiObject.PodStore) (string, error) {
	// TODO: 从容器管理器中启动一个pod中的所有容器
	retErr := ""
	for _, container := range pod.Spec.Containers {
		_, err := r.restartPodContainer(pod, &container)
		if err != nil {
			retErr += err.Error() + "\n"
			k8log.ErrorLog("kubelet", fmt.Sprintf("restart pod %s/%s failed, err %s", pod.Metadata.Namespace, pod.Metadata.Name, err.Error()))
		}
	}

	if retErr != "" {
		return "", fmt.Errorf(retErr)
	}

	k8log.InfoLog("kubelet", fmt.Sprintf("restart pod %s/%s success", pod.Metadata.Namespace, pod.Metadata.Name))
	return pod.Metadata.UUID, nil
}

// **********************************************************************************
//  单个容器的操作放在下面
// **********************************************************************************

func (r *runtimeManager) startPodContainer(pod *apiObject.PodStore, container *apiObject.Container) (string, error) {
	// TODO: 从容器管理器中启动一个普通的容器
	filter := make(map[string][]string)

	// 根据标签过滤器，过滤出来所有的容器
	filter[minik8sTypes.ContainerLabel_PodName] = []string{pod.Metadata.Name}
	filter[minik8sTypes.ContainerLabel_PodNamespace] = []string{pod.Metadata.Namespace}
	filter[minik8sTypes.ContainerLabel_PodUID] = []string{pod.Metadata.UUID}
	filter[minik8sTypes.ContainerLabel_IfPause] = []string{minik8sTypes.ContainerLabel_IfPause_False}

	// 根据容器的名字过滤器，过滤出来所有的容器
	res, err := r.containerManager.ListContainersWithOpt(filter)

	if err != nil {
		return "", err
	}

	// 遍历所有的容器，然后启动
	for _, container := range res {
		_, err := r.containerManager.StartContainer(container.ID)
		if err != nil {
			return "", err
		}
	}

	return "", nil
}

func (r *runtimeManager) stopPodContainer(pod *apiObject.PodStore, container *apiObject.Container) (string, error) {
	// TODO: 从容器管理器中停止一个普通的容器
	filter := make(map[string][]string)

	// 根据标签过滤器，过滤出来所有的容器
	filter[minik8sTypes.ContainerLabel_PodName] = []string{pod.Metadata.Name}
	filter[minik8sTypes.ContainerLabel_PodNamespace] = []string{pod.Metadata.Namespace}
	filter[minik8sTypes.ContainerLabel_PodUID] = []string{pod.Metadata.UUID}
	filter[minik8sTypes.ContainerLabel_IfPause] = []string{minik8sTypes.ContainerLabel_IfPause_False}

	// 根据容器的名字过滤器，过滤出来所有的容器
	res, err := r.containerManager.ListContainersWithOpt(filter)

	if err != nil {
		return "", err
	}

	// 遍历所有的容器，然后删除
	for _, container := range res {
		_, err := r.containerManager.StopContainer(container.ID)
		if err != nil {
			return "", err
		}
	}

	return "", nil
}

// 创建Pod里面的单个容器， 返回容器的ID
func (r *runtimeManager) createPodContainer(pod *apiObject.PodStore, container *apiObject.Container, pauseContainerID string) (string, error) {
	// TODO: 从容器管理器中创建一个普通的容器

	// [1] 拉取镜像
	// 创建一个minik8stypes.ImagePullPolicy
	imagePullPolicy := minik8sTypes.ImagePullPolicy(container.ImagePullPolicy)
	// 根据镜像的拉取策略，拉取镜像
	_, err := r.imageManager.PullImageWithPolicy(container.Image, imagePullPolicy)
	if err != nil {
		return "", err
	}

	// [2] 创建容器前，获取容器的配置
	containerConfig, err := r.getPodContainerConfig(pod, container, pauseContainerID)
	if err != nil {
		return "", err
	}

	// [3] 创建容器

	if container.Name == "" {
		container.Name = RegularContainerNameBase + uuid.NewUUID()
	}

	ID, err := r.containerManager.CreateContainer(container.Name, containerConfig)

	if err != nil {
		return "", err
	}

	// [4] 启动容器
	_, err = r.containerManager.StartContainer(ID)

	if err != nil {
		return "", err
	}

	k8log.InfoLog("kubelet", fmt.Sprintf("create Pod Container %s success, ID is %s", container.Name, ID))
	return ID, nil
}

func (r *runtimeManager) removePodContainer(pod *apiObject.PodStore, container *apiObject.Container) (string, error) {
	// TODO: 从容器管理器中删除一个普通的容器
	// 筛选出来所有标签和当前POD相关的容器
	filter := make(map[string][]string)

	// 根据标签过滤器，过滤出来所有的容器
	filter[minik8sTypes.ContainerLabel_PodName] = []string{pod.Metadata.Name}
	filter[minik8sTypes.ContainerLabel_PodNamespace] = []string{pod.Metadata.Namespace}
	filter[minik8sTypes.ContainerLabel_PodUID] = []string{pod.Metadata.UUID}
	filter[minik8sTypes.ContainerLabel_IfPause] = []string{minik8sTypes.ContainerLabel_IfPause_False}

	// 根据容器的名字过滤器，过滤出来所有的容器
	res, err := r.containerManager.ListContainersWithOpt(filter)

	if err != nil {
		return "", err
	}

	// 遍历所有的容器，然后删除
	for _, container := range res {
		_, err := r.containerManager.RemoveContainer(container.ID)
		if err != nil {
			return "", err
		}
	}

	return "", nil
}

// restartPodContainer
// 重启Pod里面的一个Container
func (r *runtimeManager) restartPodContainer(pod *apiObject.PodStore, container *apiObject.Container) (string, error) {
	// TODO: 从容器管理器中重启一个普通的容器
	// 筛选出来所有标签和当前POD相关的容器
	filter := make(map[string][]string)

	// 根据标签过滤器，过滤出来所有的容器
	filter[minik8sTypes.ContainerLabel_PodName] = []string{pod.Metadata.Name}
	filter[minik8sTypes.ContainerLabel_PodNamespace] = []string{pod.Metadata.Namespace}
	filter[minik8sTypes.ContainerLabel_PodUID] = []string{pod.Metadata.UUID}
	filter[minik8sTypes.ContainerLabel_IfPause] = []string{minik8sTypes.ContainerLabel_IfPause_False}

	// 根据容器的名字过滤器，过滤出来所有的容器
	res, err := r.containerManager.ListContainersWithOpt(filter)

	if err != nil {
		return "", err
	}

	// 遍历所有的容器，然后重启
	for _, container := range res {
		_, err := r.containerManager.RestartContainer(container.ID)
		if err != nil {
			return "", err
		}
	}

	return "", nil
}

// ******************************************************************
// 以下是一些辅助函数
// ******************************************************************

func (r *runtimeManager) parseVolumeBinds(podVolumes []apiObject.Volume, containerVolumeMounts []apiObject.VolumeMount) ([]string, error) {
	// 我们知道Pod的配置文件里面Pod有自己的Volume，然后Container可以通过Pod挂在的Volume的名字来挂载Pod的Volume
	// 所以我们需要把Pod的Volume和Container的VolumeMounts做一个映射，然后把Pod的Volume挂载到Container的VolumeMounts上面

	// 创建一个Map解析pod的volume, 先把Pod级别的所有Mounts都放到Map中
	// volumes := make(map[string]*apiObject.Volume)
	// 创建一个map，把string映射到*apiObject.Volume
	volumes := make(map[string]*apiObject.Volume)

	// 创建一个空的返回结果
	volumeBinds := []string{}

	// 遍历pod的volume，将pod的volume添加到volumes中
	for _, volume := range podVolumes {
		// 如果是hostPath类型的volume，那么就添加到volumes中
		if volume.HostPath.Path != "" {
			volumes[volume.Name] = &volume
		}
	}

	// 遍历container的volumeMounts，将container的volumeMounts添加到volumeBinds中
	for _, volumeMount := range containerVolumeMounts {
		// 需要手动的检查volumeMount.Name是否存在，如果不存在，那么就报错
		if _, ok := volumes[volumeMount.Name]; !ok {
			return nil, fmt.Errorf("volumeMount.Name %s not found in pod volumes", volumeMount.Name)
		}
		volumesValue := volumes[volumeMount.Name]

		if volumesValue.HostPath.Path == "" {
			return nil, fmt.Errorf("volumesValue.HostPath %s is not hostPath type", volumesValue.HostPath)
		}

		// volumesValue.HostPath.Pat
		// 如果存在，那么就把volumeMount.Name和volumeMount.MountPath拼接成一个字符串，添加到volumeBinds中
		volumeBindValue := volumesValue.HostPath.Path + ":" + volumeMount.MountPath

		// 简单处理，就直接把映射的字符串添加到volumeBinds中
		volumeBinds = append(volumeBinds, volumeBindValue)
	}

	return volumeBinds, nil
}

// 创建一个Pod里面的Contianer需要的配置，通过这个函数返回
func (r *runtimeManager) getPodContainerConfig(pod *apiObject.PodStore, container *apiObject.Container, pauseContainerID string) (*minik8sTypes.ContainerConfig, error) {
	// [容器标签] 为pause容器添加标签
	// 遍历pod的标签，将pod的标签添加到pause容器的标签中
	pauseLabels := map[string]string{}
	for labelKey, labelVal := range pod.Metadata.Labels {
		pauseLabels[labelKey] = labelVal
	}
	// [元标签] 为pause容器添加标签，标签的key为"pod"，value为pod的名字
	// 四个标签
	// podName、podUID、ifPause、namespace
	pauseLabels[minik8sTypes.ContainerLabel_PodName] = pod.Metadata.Name
	pauseLabels[minik8sTypes.ContainerLabel_PodUID] = string(pod.Metadata.UUID)
	pauseLabels[minik8sTypes.ContainerLabel_IfPause] = minik8sTypes.ContainerLabel_IfPause_False
	pauseLabels[minik8sTypes.ContainerLabel_PodNamespace] = pod.Metadata.Namespace

	// [环境变量] 处理好传入配置的container的环境变量和创建容器的环境变量的映射
	var containerEnv []string
	for _, env := range container.Env {
		containerEnv = append(containerEnv, env.Name+"="+env.Value)
	}

	// [PauseRef]
	pauseRef := ContianerREfPrefix + pauseContainerID
	pauseName := PauseContainerNameBase + pod.Metadata.UUID

	// [Binds] 处理好传入配置的container的volumeMounts和创建容器的volumeMounts的映射
	contianerBinds, err := r.parseVolumeBinds(pod.Spec.Volumes, container.VolumeMounts)

	if err != nil {
		return nil, err
	}

	config := minik8sTypes.ContainerConfig{
		Image:      container.Image,
		Entrypoint: container.Command,
		Cmd:        container.Args,
		Env:        containerEnv,
		Labels:     pauseLabels,
		Tty:        container.TTY,
		// 把下面的几个模式都设置为pause容器的id，才能通讯
		PidMode:     pauseRef,
		IpcMode:     pauseRef,
		NetworkMode: pauseRef,
		// 挂载容器的处理，绑定到Pause容器上面
		Volumes:     nil,
		Binds:       contianerBinds,
		VolumesFrom: []string{pauseName},
		// 资源限制
		// CPUResourceLimit: container.Resources.Limits.CPU,
		// 转int64
		CPUResourceLimit: int64(container.Resources.Limits.CPU),
		MemoryLimit:      int64(container.Resources.Limits.Memory),
	}
	return &config, nil
}
