package container

// 测试创建容器的方法
import (
	minik8stypes "miniK8s/pkg/minik8sTypes"
	"testing"
)

func TestCreateContainer(t *testing.T) {
	// 创建一个容器管理器对象
	cm := &ContainerManager{}
	// 创建一个容器的配置对象
	option := &minik8stypes.ContainerConfig{
		Image:           "nginx:latest",
		Env:             []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
		ImagePullPolicy: minik8stypes.PullAlways,
	}

	ID, err := cm.CreateContainer("nginx", option)
	if err != nil {
		t.Error(err)
	}
	println(ID)

	result, err := cm.StartContainer(ID)
	if err != nil {
		t.Error(err)
	}
	println(result)
}
