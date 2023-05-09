package runtime

import (
	"miniK8s/pkg/apiObject"
	"testing"
)

var testPod = apiObject.PodStore{
	Basic: apiObject.Basic{
		APIVersion: "v1",
		Kind:       "Pod",
		Metadata: apiObject.Metadata{
			Name:      "testPod",
			Namespace: "testNamespace",
			UUID:      "1f3a54a3-c1b9-4e47-b063-2a6d84fde222",
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

func TestCreatePod(t *testing.T) {
	// 创建一个runtimeManager
	r := NewRuntimeManager()

	// 创建pod
	err := r.CreatePod(&testPod)

	if err != nil {
		t.Error(err)
	}
}

func TestStopPod(t *testing.T) {
	// 创建一个runtimeManager
	r := NewRuntimeManager()

	// 停止pod
	err := r.StopPod(&testPod)

	if err != nil {
		t.Error(err)
	}
}

func TestStartPod(t *testing.T) {
	// 创建一个runtimeManager
	r := NewRuntimeManager()

	// 启动pod
	err := r.StartPod(&testPod)

	if err != nil {
		t.Error(err)
	}
}

func TestRestartPod(t *testing.T) {
	// 创建一个runtimeManager
	r := NewRuntimeManager()

	// 重启pod
	err := r.RestartPod(&testPod)

	if err != nil {
		t.Error(err)
	}
}

func TestDeletePod(t *testing.T) {
	// 创建一个runtimeManager
	r := NewRuntimeManager()

	// 删除pod
	err := r.DeletePod(&testPod)

	if err != nil {
		t.Error(err)
	}
}

// Spec: apiObject.PodSpec{
// 	Containers: []apiObject.Container{},
// },
// Name:      "testPod",
// 	Namespace: "testNamespace",
// 	UUID:      "1f3a54a3-c1b9-4e47-b063-2a6d84fde222",
