package worker

import (
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/message"
	"miniK8s/util/uuid"
	"testing"
	"time"
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

var workerManager = NewPodWorkerManager()

// 以下所有测试存在顺序关系，请务必按照顺序依次执行
func TestAddPod(t *testing.T) {
	// 添加 Pod
	podUpdate := &entity.PodUpdate{
		Action:    message.CREATE,
		PodTarget: testPod,
		Node:      "testNode",
	}
	err := workerManager.AddPod(&podUpdate.PodTarget)
	if err != nil {
		if err.Error() != "pod already exists" {
			t.Errorf("AddPod error: %v", err)
		}
	}

	// 检查 Pod 是否存在
	// if _, ok := workerManager.PodWorkersMap[testPod.GetPodUUID()]; !ok {
	// 	t.Errorf("Pod not added to PodWorkerManager")
	// }

}

func TestStartPod(t *testing.T) {
	// 启动 Pod
	err := workerManager.StartPod(&testPod)
	if err != nil {
		t.Errorf("StartPod error: %v", err)
	}
	// 检查 Pod 是否存在
	// if _, ok := workerManager.PodWorkersMap[testPod.GetPodUUID()]; !ok {
	// 	t.Errorf("Pod not started")
	// }
}

func TestStopPod(t *testing.T) {
	// 等待容器启动
	time.Sleep(5 * time.Second)
	// 停止 Pod
	// err := workerManager.StopPod(&testPod)
	// if err != nil {
	// 	t.Errorf("StopPod error: %v", err)
	// }
}

func TestRestartPod(t *testing.T) {
	// 重启 Pod
	err := workerManager.RestartPod(&testPod)
	if err != nil {
		t.Errorf("RestartPod error: %v", err)
	}
	// 检查 Pod 是否存在
	// if _, ok := workerManager.PodWorkersMap[testPod.GetPodUUID()]; !ok {
	// 	t.Errorf("Pod not restarted")
	// }
}

func TestDeletePod(t *testing.T) {

	// 删除 Pod
	err := workerManager.DeletePod(&testPod)
	if err != nil {
		t.Errorf("DeletePod error: %v", err)
	}
	// 检查 Pod 是否存在
	// if _, ok := workerManager.PodWorkersMap[testPod.GetPodUUID()]; ok {
	// 	t.Errorf("Pod not deleted from PodWorkerManager")
	// }
}
