package proxy

import (
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/message"
	"os"
	"testing"
)

var testService = apiObject.ServiceStore{
	Basic: apiObject.Basic{
		APIVersion: "v1",
		Kind:       "Service",
		Metadata: apiObject.Metadata{
			Name:      "testService",
			Namespace: "testNamespace",
			UUID:      "1f3a54a3-c1b9-4e47-b063-2a6d84fde222",
		},
	},
	Spec: apiObject.ServiceSpec{
		Selector: map[string]string{
			"app": "test",
		},
		Ports: []apiObject.ServicePort{
			{
				Port:       80,
				TargetPort: 80,
				Name:       "testService",
			},
		},
	},
}

func TestSaveIPTables(t *testing.T) {
	im := NewIptableManager()
	path := "test-save-iptables"

	// Ensure file does not exist
	if _, err := os.Stat(path); err == nil {
		os.Remove(path)
	}

	// Save iptables
	err := im.SaveIPTables(path)
	if err != nil {
		t.Errorf("failed to save iptables: %v", err)
	}

	// Ensure file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("file was not created: %v", err)
	}

	// Cleanup
	// os.Remove(path)
}

func TestDeleteService(t *testing.T) {
	im := NewIptableManager()
	serviceUpdate := &entity.ServiceUpdate{
		Action:        message.DELETE,
		ServiceTarget: testService,
	}
	// im.CreateService(serviceUpdate)

	err := im.DeleteService(serviceUpdate)
	if err != nil {
		t.Error(err)
	}
}

// func TestRestoreIPTables(t *testing.T) {
// 	im := New()
// 	path := "test-restore-iptables"

// 	// Ensure file does not exist
// 	if _, err := os.Stat(path); err == nil {
// 		os.Remove(path)
// 	}

// 	// Create test iptables file
// 	err := ioutil.WriteFile(path, []byte("*filter\n:INPUT ACCEPT [0:0]\n:FORWARD ACCEPT [0:0]\n:OUTPUT ACCEPT [0:0]\nCOMMIT\n"), 0644)
// 	if err != nil {
// 		t.Errorf("failed to create test iptables file: %v", err)
// 	}

// 	// Restore iptables
// 	err = im.RestoreIPTables(path)
// 	if err != nil {
// 		t.Errorf("failed to restore iptables: %v", err)
// 	}

// 	// Cleanup
// 	os.Remove(path)
// }

// func TestCreateService(t *testing.T) {
// 	// 运行 kube-proxy，监听service管道的变化
// 	proxy := NewKubeProxy(listwatcher.DefaultListwatcherConfig())
// 	go proxy.Run()

// 	gin.SetMode(gin.ReleaseMode)
// 	// 创建一个新的gin引擎，并注册AddService处理函数。
// 	r := gin.New()
// 	// 关闭gin的日志输出
// 	r.Use(gin.LoggerWithWriter(io.Discard))
// 	// 设置gin为生产模式
// 	gin.SetMode(gin.ReleaseMode)
// 	// 通过调用gin引擎的ServeHTTP方法，可以模拟一个http请求，从而测试AddService方法。
// 	r.POST(config.ServiceURL, handlers.AddService)

// 	// 读取文件"./testFile/yamlFile/Service-i.yaml"，将文件内容作为请求体。
// 	// 打开文件

// 	file, err := os.Open("../testFile/Service.yaml")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	// 读取文件内容
// 	content, err := io.ReadAll(file)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// 将文件内容转换为Node对象
// 	service := &apiObject.Service{}
// 	err = yaml.Unmarshal(content, service)

// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	serviceUpdate := &entity.ServiceUpdate{
// 		Action:        entity.CREATE,
// 		ServiceTarget: apiObject.ServiceStore{},
// 	}

// // 读取的内容转化为jsons
// jsonBytes, err := json.Marshal(service)

// if err != nil {
// 	t.Fatal(err)
// }
// serviceReader := bytes.NewReader(jsonBytes)

// URL := config.ServiceURL
// URL = stringutil.Replace(URL, config.URL_PARAM_NAMESPACE_PART, service.GetNamespace())
// t.Log("request url:" + URL)
// // 创建一个http请求，请求方法为POST，请求路径为"/api/v1/services"，请求体为一个json字符串。
// req, err := http.NewRequest("POST", URL, serviceReader)

// if err != nil {
// 	t.Fatal(err)
// }
// req.Header.Set("Content-Type", "application/json")

// // 创建响应写入器
// w := httptest.NewRecorder()

// // 将请求和响应写入gin.Context
// c, _ := gin.CreateTestContext(w)
// c.Request = req

// // 执行处理函数
// r.HandleContext(c)

// // 获取响应结果
// resp := w.Result()

// if resp.StatusCode != http.StatusCreated {
// 	t.Errorf("expected status %v but got %v", http.StatusOK, resp.StatusCode)
// }

// 	proxy.iptableManager.CreateService(serviceUpdate)
// }

// func TestClearIPTables(t *testing.T) {
// 	// 创建 IptableManager 实例
// 	im := New()

// 	// 添加一些规则
// 	im.ipt.Append("filter", "INPUT", "-s 127.0.0.1/32 -p tcp -m tcp --dport 22 -j ACCEPT")
// 	im.ipt.Append("filter", "INPUT", "-s 192.168.1.0/24 -p tcp -m tcp --dport 80 -j ACCEPT")

// 	// 清除规则
// 	im.ClearIPTables()

// 	// 检查是否已清除所有规则
// 	chains, _ := im.ipt.List("filter", "INPUT")
// 	if (len(chains)) != 0 {
// 		t.Errorf("ClearIPTables failed, iptables chains are not empty")
// 	}

// 	im.SaveIPTables("test-save-iptables")
// }
