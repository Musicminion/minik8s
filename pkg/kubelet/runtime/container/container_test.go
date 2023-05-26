package container

import (
	"testing"
)

var TestContainerURLs = []string{
	"docker.io/library/nginx:latest",
	"docker.io/library/redis:latest",
}

// 遍历启动所有的容器
var opt = map[string][]string{
	"test": {"test"},
}

// 测试之前执行的方法
// func TestMain(m *testing.M) {
// 	// 先列出所有的容器，然后删除所有的容器
// 	cm := &ContainerManager{}

// 	containers, err := cm.ListContainersWithOpt(opt)
// 	if err != nil {
// 		panic(err)
// 	}
// 	for _, container := range containers {
// 		_, err := cm.RemoveContainer(container.ID)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// 	m.Run()
// }

// 测试创建容器的方法

// func TestCreateContainer(t *testing.T) {
// 	// 创建一个容器管理器对象
// 	cm := &ContainerManager{}

// 	// 依次创建容器
// 	for id, url := range TestContainerURLs {
// 		// 创建一个容器的配置对象
// 		option := &minik8stypes.ContainerConfig{
// 			Image:           url,
// 			Env:             []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
// 			ImagePullPolicy: minik8stypes.PullIfNotPresent,
// 			Labels:          map[string]string{"test": "test"},
// 		}
// 		containerName := "test" + strconv.Itoa(id)
// 		ID, err := cm.CreateContainer(containerName, option)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		cm.StartContainer(ID)
// 		t.Logf("[Created] Container ID: %s", ID)
// 	}
// }

// // 测试列出所有的容器的方法
// func TestListContainers(t *testing.T) {
// 	// 创建一个容器管理器对象
// 	cm := &ContainerManager{}
// 	// 调用列出所有容器的方法
// 	containers, err := cm.ListContainersWithOpt(opt)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	// 遍历打印所有容器的信息
// 	for _, container := range containers {
// 		t.Logf("Container ID: %s, Container Name: %s", container.ID, container.Names)
// 	}
// }

// // 测试获取运行中容器的状态的方法
// func TestGetContainerStats(t *testing.T) {
// 	// 创建一个容器管理器对象
// 	cm := &ContainerManager{}
// 	// 调用列出所有容器的方法
// 	containers, err := cm.ListContainersWithOpt(opt)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	// 依次获取每个容器的状态,比如CPU使用率
// 	for _, container := range containers {
// 		status, err := cm.GetContainerStats(container.ID)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		t.Logf("Container %s: MemoryStats: %d", container.ID, status.MemoryStats.Usage)

// 		t.Logf("Container %s: CPUStats: %d", container.ID, status.CPUStats.CPUUsage.TotalUsage)
// 	}
// }

// // 测试GetContainerInspectInfo
// func TestGetContainerInspectInfo(t *testing.T) {
// 	// 创建一个容器管理器对象
// 	cm := &ContainerManager{}
// 	// 遍历获取所有容器的信息
// 	containers, err := cm.ListContainersWithOpt(opt)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	for _, container := range containers {
// 		info, err := cm.GetContainerInspectInfo(container.ID)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		t.Logf("Container ID: %s, Container Name: %s", info.ID, info.Name)
// 	}
// }

// // 测试停止容器的方法
// func TestStopContainer(t *testing.T) {
// 	// 创建一个容器管理器对象
// 	cm := &ContainerManager{}
// 	// 遍历停止所有的容器
// 	containers, err := cm.ListContainersWithOpt(opt)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	for _, container := range containers {
// 		_, err := cm.StopContainer(container.ID)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 	}
// }

// // 测试启动容器的方法
// func TestStartContainer(t *testing.T) {
// 	// 创建一个容器管理器对象
// 	cm := &ContainerManager{}
// 	// 遍历启动所有的容器
// 	opt := map[string][]string{}
// 	opt["test"] = []string{"test"}

// 	containers, err := cm.ListContainersWithOpt(opt)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	for _, container := range containers {
// 		_, err := cm.StartContainer(container.ID)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 	}
// }

// // func TestExecContainer(t *testing.T) {
// // 	// 定义测试用的容器 ID 和命令
// // 	cmd := []string{"sh", "-c", "touch /testfile"}
// // 	// cmd := []string{"touch", "new"}
// // 	// 创建一个 ContainerManager 实例
// // 	cm := &ContainerManager{}
// // 	containers, err := cm.ListContainersWithOpt(opt)
// // 	if err != nil {
// // 		t.Error(err)
// // 	}

// // 	for _, container := range containers {
// // 		out, err := cm.ExecContainer(container.ID, cmd)
// // 		if err != nil {
// // 			t.Error(err)
// // 		}
// // 		k8log.DebugLog("Container Manager", "out is "+string(out))
// // 	}
// // }

// // 测试删除容器的方法
// // func TestRemoveContainer(t *testing.T) {
// // 	// 创建一个容器管理器对象
// // 	cm := &ContainerManager{}
// // 	// 遍历删除所有的容器
// // 	containers, err := cm.ListContainersWithOpt(opt)
// // 	if err != nil {
// // 		t.Error(err)
// // 	}
// // 	for _, container := range containers {
// // 		_, err := cm.RemoveContainer(container.ID)
// // 		if err != nil {
// // 			t.Error(err)
// // 		}
// // 	}
// // }

func TestGetContainerResource(t *testing.T) {
	// 创建一个容器管理器对象
	cm := &ContainerManager{}
	// 遍历获取所有容器的信息
	opt := map[string][]string{}
	containers, err := cm.ListContainersWithOpt(opt)
	if err != nil {
		t.Error(err)
	}

	// 对每个container，计算cpu和memory的使用率
	for _, container := range containers {
		stats, err := cm.GetContainerStats(container.ID)
		if err != nil {
			t.Error(err)
		}
		// 计算cpu使用率
		cpuPercent := calculateCPUPercentUnix(stats)
		t.Logf("Container %s: CPU Usage: %f", container.ID, cpuPercent)
		// 计算memory使用率
		memoryPercent := calculateMemoryPercentUnix(stats)
		t.Logf("Container %s: Memory Usage: %f", container.ID, memoryPercent)
	}
}
