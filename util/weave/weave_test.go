package weave

import (
	"testing"
)

func TestWeaveFindIpByContainerID(t *testing.T) {
	// // 假设Docker容器的ID为test_container，IP地址为
	// containerID := "63730fbf5ce8"
	// res, err := WeaveFindIpByContainerID(containerID)

	// if err != nil {
	// 	t.Errorf("WeaveFindIpByContainerID() error = %v", err)
	// 	return
	// }

	// t.Logf("WeaveFindIpByContainerID() result = %v", res)
}

func TestWeaveAttach(t *testing.T) {
	// 假设Docker容器的ID为test_container，IP地址为192.168.0.2
	// containerID := "test1"
	// containerIP := "192.168.0.2"

	// 假设Weave命令返回的IP地址与容器IP相同
	// wantIP := "10.244.0.1"

	// // 使用模拟环境运行测试
	// weaveIP, err := testutil.SetupTestNetwork(containerID, containerIP)
	// if err != nil {
	// 	t.Errorf("setupTestNetwork() error = %v", err)
	// 	return
	// }
	// defer func() {
	// 	// 清理模拟环境
	// 	err = testutil.CleanupTestNetwork(containerID)
	// 	if err != nil {
	// 		t.Errorf("cleanupTestNetwork() error = %v", err)
	// 	}
	// }()

	// 调用WeaveAttach()函数
	// gotIP, err := WeaveAttach(containerID)
	// if err != nil {
	// 	t.Errorf("WeaveAttach(%q, %q) error = %v", containerID, wantIP, err)
	// }
	// if gotIP != wantIP {
	// 	t.Errorf("WeaveAttach(%q, %q) = %q, want %q", containerID, wantIP, gotIP, wantIP)
	// }
}
