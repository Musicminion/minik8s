package runtime

import (
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/kubelet/runtime/container"
	minik8stypes "miniK8s/pkg/minik8sTypes"
	"miniK8s/util/uuid"
	"testing"
)

var testPod = apiObject.PodStore{
	Basic: apiObject.Basic{
		APIVersion: "v1",
		Kind:       "Pod",
		Metadata: apiObject.Metadata{
			Name:      "testPod",
			Namespace: "testNamespace",
			UUID:      uuid.NewUUID(),
			Labels: map[string]string{
				"app": "test",
			},
		},
	},
	Spec: apiObject.PodSpec{
		Containers: []apiObject.Container{
			{
				Name:  "testContainer-1",
				Image: "docker.io/library/redis",
			},
			{
				Name:  "testContainer-2",
				Image: "docker.io/library/nginx",
			},
		},
	},
}

func TestMain(m *testing.M) {
	// 初始化containerManager
	containerManager := container.ContainerManager{}
	result, err := containerManager.ListContainers()

	if err != nil {
		panic(err)
	}

	// 遍历所有容器
	for _, container := range result {
		if container.Labels[minik8stypes.ContainerLabel_PodName] == "testPod" && container.Labels[minik8stypes.ContainerLabel_PodNamespace] == "testNamespace" {
			// 删除容器
			_, err := containerManager.RemoveContainer(container.ID)
			if err != nil {
				panic(err)
			}
		}
	}

	result, err = containerManager.ListContainers()
	if err != nil {
		panic(err)
	}

	// 遍历所有容器
	for _, container := range result {
		// 遍历 container.Names
		for _, name := range container.Names {
			if name == "testContainer-1" || name == "testContainer-2" {
				// 删除容器
				_, err := containerManager.RemoveContainer(container.ID)
				if err != nil {
					panic(err)
				}
			}
		}
	}

	m.Run()
}

// func TestCreatePod(t *testing.T) {
// 	// 创建一个runtimeManager

// 	r := NewRuntimeManager()
// 	err := r.DeletePod(&testPod)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	// 创建pod
// 	err = r.CreatePod(&testPod)

// 	if err != nil {
// 		t.Error(err)
// 	}

// }

// func TestStopPod(t *testing.T) {
// 	// 创建一个runtimeManager
// 	r := NewRuntimeManager()

// 	// 停止pod
// 	err := r.StopPod(&testPod)

// 	if err != nil {
// 		t.Error(err)
// 	}
// }

// func TestStartPod(t *testing.T) {
// 	// 创建一个runtimeManager
// 	r := NewRuntimeManager()

// 	// 启动pod
// 	err := r.StartPod(&testPod)

// 	if err != nil {
// 		t.Error(err)
// 	}
// }

// func TestRestartPod(t *testing.T) {
// 	// 创建一个runtimeManager
// 	r := NewRuntimeManager()

// 	// 重启pod
// 	err := r.RestartPod(&testPod)

// 	if err != nil {
// 		t.Error(err)
// 	}
// }

// func TestDeletePod(t *testing.T) {
// 	// 创建一个runtimeManager
// 	r := NewRuntimeManager()

// 	// 删除pod
// 	err := r.DeletePod(&testPod)

// 	if err != nil {
// 		t.Error(err)
// 	}
// }

// func TestCleanAll(t *testing.T) {
// 	// 初始化containerManager
// 	containerManager := container.ContainerManager{}
// 	result, err := containerManager.ListContainers()

// 	if err != nil {
// 		panic(err)
// 	}

// 	// 遍历所有容器
// 	for _, container := range result {
// 		if container.Labels[minik8stypes.ContainerLabel_PodName] == "testPod" && container.Labels[minik8stypes.ContainerLabel_PodNamespace] == "testNamespace" {
// 			// 删除容器
// 			_, err := containerManager.RemoveContainer(container.ID)
// 			if err != nil {
// 				panic(err)
// 			}
// 		}
// 	}
// }

// func TestRecreatePodContainer(t *testing.T) {
// 	// 创建一个runtimeManager
// 	r := NewRuntimeManager()
// 	containerManager := container.ContainerManager{}

// 	// 创建pod
// 	err := r.CreatePod(&testPod)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	t.Log("create pod name " + testPod.GetPodName())
// 	// 启动pod
// 	err = r.StartPod(&testPod)
// 	t.Log("start pod name " + testPod.GetPodName())

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	filter := make(map[string][]string)
// 	filter[minik8sTypes.ContainerLabel_PodUID] = []string{testPod.GetPodUUID()}

// 	containers, err := containerManager.ListContainersWithOpt(filter)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	// 删除容器
// 	t.Log("remove container name " + containers[0].Names[0])
// 	containerManager.RemoveContainer(containers[0].ID)

// 	// 重启pod的容器
// 	r.RecreatePodContainer(&testPod)
// 	t.Log("recreate pod name " + testPod.GetPodName())

// 	if err != nil {
// 		t.Error(err)
// 	}

// }

// func TestCreatePodAndSaveToEtcd(t *testing.T) {
// 	// 创建一个runtimeManager
// 	r := NewRuntimeManager()

// 	r.DeletePod(&testPod)
// 	// 创建pod
// 	err := r.CreatePod(&testPod)

// 	if err != nil {
// 		t.Error(err)
// 	}
// 	// 把PodStore转化为json
// 	podStoreJson, err := json.Marshal(testPod)
// 	if err != nil {
// 		return
// 	}

// 	// 将pod存储到etcd中
// 	// 持久化
// 	// key = stringutil.Replace(serverconfig.DefaultPod, config.URI_PARAM_NAME_PART, newPodName)

// 	key := fmt.Sprintf(serverconfig.EtcdPodPath+"%s/%s", testPod.GetPodNamespace(), testPod.GetPodName())

// 	// 将pod存储到etcd中
// 	err = etcdclient.EtcdStore.Put(key, podStoreJson)

// 	if err != nil {
// 		t.Error(err)
// 	}
// 	etcdclient.EtcdStore.Put(key, podStoreJson)

// 	// 创建一个容器管理器对象
// 	cm := &container.ContainerManager{}
// 	var opt = map[string][]string{
// 		"test": {"test"},
// 	}
// 	containers, err := cm.ListContainersWithOpt(opt)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	for _, container := range containers {
// 		_, err := cm.RemoveContainer(container.ID)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 	}
// }

// Spec: apiObject.PodSpec{
// 	Containers: []apiObject.Container{},
// },
// Name:      "testPod",
// 	Namespace: "testNamespace",
// 	UUID:      "1f3a54a3-c1b9-4e47-b063-2a6d84fde222",

// func TestCreatePodAndSaveToEtcd(t *testing.T) {
// 	// 创建一个runtimeManager
// 	r := NewRuntimeManager()

// 	r.DeletePod(&testPod)
// 	// 创建pod
// 	err := r.CreatePod(&testPod)

// 	if err != nil {
// 		t.Error(err)
// 	}
// 	// 把PodStore转化为json
// 	podStoreJson, err := json.Marshal(testPod)
// 	if err != nil {
// 		return
// 	}

// 	// 将pod存储到etcd中
// 	// 持久化
// 	// key = stringutil.Replace(serverconfig.DefaultPod, config.URI_PARAM_NAME_PART, newPodName)

// 	key := fmt.Sprintf(serverconfig.EtcdPodPath+"%s/%s", testPod.GetPodNamespace(), testPod.GetPodName())

// 	// 将pod存储到etcd中
// 	err = etcdclient.EtcdStore.Put(key, podStoreJson)

// 	if err != nil {
// 		t.Error(err)
// 	}
// 	etcdclient.EtcdStore.Put(key, podStoreJson)

// 	// 创建一个容器管理器对象
// 	cm := &container.ContainerManager{}
// 	var opt = map[string][]string{
// 		"test": {"test"},
// 	}
// 	containers, err := cm.ListContainersWithOpt(opt)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	for _, container := range containers {
// 		_, err := cm.RemoveContainer(container.ID)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 	}
// }
