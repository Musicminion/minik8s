package nginx

import (
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/util/file"
	"os"
	"testing"
	// "github.com/stretchr/testify/assert"
)

var dns = apiObject.Dns{
	Spec: apiObject.DnsSpec{
		Host: "example.com",
		Paths: []apiObject.Path{
			{
				SubPath: "/api/v1",
				SvcName: "example-service1",
				SvcPort: "80",
				SvcIp:   "192.168.1.1",
			},
			{
				SubPath: "/api/v2",
				SvcName: "example-service2",
				SvcPort: "8080",
				SvcIp:   "192.168.1.2",
			},
		},
	},
}

func TestFormatConf(t *testing.T) {
	// 创建一个Dns对象，用于测试

	// 定义 FormatConf 函数应该返回的字符串
	expected := "# example.com.conf\nserver {\n\tlisten 80;\n\tserver_name example.com;\n\tlocation /api/v1 {\n\t\tproxy_pass http://192.168.1.1:80/;\n\t}\n\tlocation /api/v2 {\n\t\tproxy_pass http://192.168.1.2:8080/;\n\t}\n}"

	// 调用 FormatConf 函数生成实际的字符串
	actual := FormatConf(dns)

	// 使用 assert 包来比较期望的字符串和实际的字符串是否相等
	if expected != actual {
		t.Errorf("expected: %s, actual: %s", expected, actual)
	}
}

func TestWriteConf(t *testing.T) {
	// 创建一个Dns对象，用于测试
	conf := FormatConf(dns)

	// 调用 WriteConf 函数
	err := WriteConf(dns, conf)
	if err != nil {
		t.Errorf("write conf failed: %s", err.Error())
	}

	// 读取文件内容
	confFileName := fmt.Sprintf("%s.conf", dns.Spec.Host)
	confFilePath := fmt.Sprintf(config.NginxConfigPath + confFileName)
	actual, err := file.ReadFile(confFilePath)
	if err != nil {
		t.Errorf("open file failed: %s", err.Error())
	}
	// 验证文件内容与写的一致
	if conf != string(actual) {
		t.Errorf("expected: %s, actual: %s", conf, string(actual))
	}

}

func TestDeleteConf(t *testing.T) {
	// 创建一个Dns对象，用于测试
	conf := FormatConf(dns)

	// 调用 WriteConf 函数
	WriteConf(dns, conf)

	// 调用 RemoveConf 函数
	DeleteConf(dns)

	// 验证文件是否被删除
	confFileName := fmt.Sprintf("%s.conf", dns.Spec.Host)
	confFilePath := fmt.Sprintf(config.NginxConfigPath + confFileName)
	_, err := os.Stat(confFilePath)
	if err == nil {
		t.Errorf("file should be removed")
	}
}
