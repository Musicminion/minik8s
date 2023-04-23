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
