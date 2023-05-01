package runtime

import (
	"miniK8s/pkg/apiObject"
	minik8sTypes "miniK8s/pkg/minik8sTypes"
	minik8stypes "miniK8s/pkg/minik8sTypes"
	"miniK8s/util/uuid"
)

func (r *runtimeManager) createPodAllContainer(pod *apiObject.PodStore) (string, error) {
	// TODO: 从容器管理器中启动一个pod中的所有容器
	for _, container := range pod.Spec.Containers {
		_, err := r.createPodContainer(pod, &container)
		if err != nil {
			// 可能需要注意垃圾回收机制
			r.removePodAllContiner(pod)
			return "", err
		}
	}

	return pod.Metadata.UUID, nil
}

func (r *runtimeManager) removePodAllContiner(pod *apiObject.PodStore) (string, error) {
	// TODO: 从容器管理器中删除一个pod中的所有容器
	for _, container := range pod.Spec.Containers {
		_, err := r.removePodContainer(pod, &container)
		if err != nil {
			return "", err
		}
	}
	return pod.Metadata.UUID, nil
}

// func (r *runtimeManager) stopPodAllContainer(pod *apiObject.PodStore) (string, error) {
// 	// TODO: 从容器管理器中删除一个pod中的所有容器
// 	for _, container := range pod.Spec.Containers {
// 		_, err := r.stopPodContainer(pod, &container)
// 		if err != nil {
// 			return "", err
// 		}
// 	}
// 	return pod.Metadata.UUID, nil
// }

// func (r *runtimeManager) startPodContainer(pod *apiObject.PodStore, container *apiObject.Container) (string, error) {
// 	// TODO: 从容器管理器中启动一个普通的容器
// 	return "", nil
// }

// func (r *runtimeManager) stopPodContainer(pod *apiObject.PodStore, container *apiObject.Container) (string, error) {
// 	// TODO: 从容器管理器中停止一个普通的容器
// 	return "", nil
// }

func (r *runtimeManager) createPodContainer(pod *apiObject.PodStore, container *apiObject.Container) (string, error) {
	// TODO: 从容器管理器中创建一个普通的容器

	// [1] 拉取镜像
	// 创建一个minik8stypes.ImagePullPolicy
	imagePullPolicy := minik8stypes.ImagePullPolicy(container.ImagePullPolicy)
	_, err := r.imageManager.PullImageWithPolicy(container.Image, imagePullPolicy)
	if err != nil {
		return "", err
	}

	// [2] 创建容器前，获取容器的配置
	containerConfig, err := r.getPodContainerConfig(pod)

	if err != nil {
		return "", err
	}

	// [3] 创建容器
	newContianerName := RegularContainerNameBase + uuid.NewUUID()
	_, err = r.containerManager.CreateContainer(newContianerName, containerConfig)

	if err != nil {
		return "", err
	}

	return newContianerName, nil
}

func (r *runtimeManager) removePodContainer(pod *apiObject.PodStore, container *apiObject.Container) (string, error) {
	// TODO: 从容器管理器中删除一个普通的容器
	return "", nil
}

func (r *runtimeManager) getPodContainerConfig(pod *apiObject.PodStore) (*minik8sTypes.ContainerConfig, error) {
	return nil, nil
}
