package status

import (
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"testing"
)

// 测试执行之前，先删除Redis中的所有数据
func TestMain(m *testing.M) {
	// 1. 删除Redis中的所有数据
	statusManager := NewStatusManager(config.GetMasterIP())
	statusManager.ResetCache()

	// 2. 执行测试用例
	m.Run()
}

func TestAddPodToCache(t *testing.T) {
	// 1. 初始化一个Pod对象
	pod1 := &apiObject.PodStore{
		Basic: apiObject.Basic{
			Kind:       "Pod",
			APIVersion: "v1",
			Metadata: apiObject.Metadata{
				Name:      "pod1",
				UUID:      "ABCD-ABCD-ABCD-ABCD",
				Namespace: "default",
			},
		},
	}
	pod2 := &apiObject.PodStore{
		Basic: apiObject.Basic{
			Kind:       "Pod",
			APIVersion: "v1",
			Metadata: apiObject.Metadata{
				Name:      "pod2",
				UUID:      "EFGH-EFGH-EFGH-EFGH",
				Namespace: "default",
			},
		},
	}

	// 2. 初始化一个StatusManager对象
	statusManager := NewStatusManager(config.GetAPIServerURLPrefix())
	err := statusManager.AddPodToCache(pod1)

	// 3. 检查返回值
	if err != nil {
		t.Errorf("AddPodToCache() error = %v", err)
	}

	err = statusManager.AddPodToCache(pod2)
	if err != nil {
		t.Errorf("AddPodToCache() error = %v", err)
	}
}

func TestGetPodFromCache(t *testing.T) {
	// 1. 初始化一个StatusManager对象
	statusManager := NewStatusManager(config.GetAPIServerURLPrefix())

	// 2. 调用GetPodFromCache()函数
	pod, err := statusManager.GetPodFromCache("ABCD-ABCD-ABCD-ABCD")
	if err != nil {
		t.Errorf("GetPodFromCache() error = %v", err)
	}

	// 3. 检查返回值
	if pod.Metadata.Name != "pod1" {
		t.Errorf("GetPodFromCache() error, pod name = %s", pod.Metadata.Name)
	}

	// 4. 再次调用GetPodFromCache()函数
	pod, err = statusManager.GetPodFromCache("EFGH-EFGH-EFGH-EFGH")
	if err != nil {
		t.Errorf("GetPodFromCache() error = %v", err)
	}

	// 5. 检查返回值
	if pod.Metadata.Name != "pod2" {
		t.Errorf("GetPodFromCache() error, pod name = %s", pod.Metadata.Name)
	}

}

func TestGetAllPodFromCache(t *testing.T) {
	// 1. 初始化一个StatusManager对象
	statusManager := NewStatusManager(config.GetAPIServerURLPrefix())

	// 2. 调用GetAllPodFromCache()函数
	podMap, err := statusManager.GetAllPodFromCache()

	// 3. 检查返回值
	if err != nil {
		t.Errorf("GetAllPodFromCache() error = %v", err)
	}

	// 4. 遍历podMap，检查pod的数量
	if len(podMap) != 2 {
		t.Errorf("GetAllPodFromCache() error, pod number = %d", len(podMap))
	}

	// 5. 遍历podMap，检查pod的名称
	for _, pod := range podMap {
		if pod.Metadata.Name != "pod1" && pod.Metadata.Name != "pod2" {
			t.Errorf("GetAllPodFromCache() error, pod name = %s", pod.Metadata.Name)
		}
	}

}

func TestDelPodFromCache(t *testing.T) {
	// 1. 初始化一个StatusManager对象
	statusManager := NewStatusManager(config.GetAPIServerURLPrefix())

	// 2. 调用DelPodFromCache()函数
	err := statusManager.DelPodFromCache("ABCD-ABCD-ABCD-ABCD")

	// 3. 检查返回值
	if err != nil {
		t.Errorf("DelPodFromCache() error = %v", err)
	}

	// 4. 再次调用GetPodFromCache()函数
	pod, err := statusManager.GetPodFromCache("ABCD-ABCD-ABCD-ABCD")

	if err != nil {
		t.Errorf("DelPodFromCache() error, pod = %v", pod)
	}

	if pod != nil {
		t.Errorf("DelPodFromCache() error, pod = %v", pod)
	}

}

func TestReset(t *testing.T) {
	// 1. 初始化一个StatusManager对象
	statusManager := NewStatusManager(config.GetAPIServerURLPrefix())

	// 2. 调用Reset()函数
	err := statusManager.ResetCache()

	// 3. 检查返回值
	if err != nil {
		t.Errorf("Reset() error = %v", err)
	}

	// 4. 再次调用GetAllPodFromCache()函数
	podMap, err := statusManager.GetAllPodFromCache()

	// 5. 检查返回值
	if err != nil {
		t.Errorf("GetAllPodFromCache() error = %v", err)
	}

	// 6. 遍历podMap，检查pod的数量
	if len(podMap) != 0 {
		t.Errorf("GetAllPodFromCache() error, pod number = %d", len(podMap))
	}

}
