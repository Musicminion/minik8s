// 测试host相关的操作

package host

import "testing"

func TestGetHostName(t *testing.T) {
	hostname := GetHostName()
	t.Log(hostname)
}

func TestGetHostIp(t *testing.T) {
	ip, err := GetHostIp()
	if err != nil {
		t.Error(err)
	} else {
		t.Log(ip)
	}

}

func TestGetSystemMemoryUsage(t *testing.T) {
	percent, err := GetHostSystemMemoryUsage()
	if err != nil {
		t.Error(err)
	} else {
		t.Log(percent)
	}
}

func TestGetSystemCpuUsage(t *testing.T) {
	percent, err := GetHostSystemCPUUsage()
	if err != nil {
		t.Error(err)
	} else {
		t.Log(percent)
	}
}
