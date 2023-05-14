package scheduler

import (
	"reflect"
	"testing"
)

// func TestGetAllNode(t *testing.T) {
// 	code, res, err := netrequest.GetRequest("http://localhost:8090/api/v1/nodes")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	if code != 200 {
// 		t.Error("code is not 200")
// 	}

// 	t.Log(res["data"])

// 	// res["data"]转化为字符串
// 	t.Log()

// 	dataStr := fmt.Sprint(res["data"])
// 	t.Log(dataStr)

// 	var nodes []apiObject.NodeStore

// 	err = json.Unmarshal([]byte(dataStr), &nodes)

// 	if err != nil {
// 		t.Error(err)
// 	}

// }

// func TestTmp(t *testing.T) {
// 	var nodes []apiObject.NodeStore

// 	code, err := netrequest.GetRequestByTarget("http://localhost:8090/api/v1/nodes", &nodes, "data")

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	if code != 200 {
// 		t.Error("code is not 200")
// 	}

// 	// 遍历nodes
// 	for _, node := range nodes {
// 		t.Log(node.GetAPIVersion())
// 	}
// }

func TestNewScheduler(t *testing.T) {
	tests := []struct {
		name    string
		want    *Scheduler
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewScheduler()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewScheduler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewScheduler() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestAddService(t *testing.T) {
// 	// var sch *Scheduler = &Scheduler{}
// 	gin.SetMode(gin.ReleaseMode)
// 	// 创建一个新的gin引擎，并注册AddService处理函数。
// 	r := gin.New()
// 	// 关闭gin的日志输出
// 	r.Use(gin.LoggerWithWriter(io.Discard))
// 	// 设置gin为生产模式
// 	gin.SetMode(gin.ReleaseMode)
// 	// 通过调用gin引擎的ServeHTTP方法，可以模拟一个http请求，从而测试AddService方法。
// 	r.POST(config.ServiceURL)

// 	// 读取文件"./testFile/yamlFile/Service-i.yaml"，将文件内容作为请求体。
// 	// 打开文件
// 	for i := 1; i <= 2; i++ {
// 		path := "./testFile/yamlFile/Service-" + fmt.Sprint(i) + ".yaml"
// 		file, err := os.Open(path)

// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		// 读取文件内容
// 		content, err := io.ReadAll(file)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		// 将文件内容转换为Service对象
// 		// 通过调用gin引擎的ServeHTTP方法，可以模拟一个http请求，从而测试AddService方法。
// 		service := &apiObject.ServiceStore{}
// 		err = yaml.Unmarshal(content, service)

// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		// 读取的内容转化为json

// 		jsonBytes, err := json.Marshal(service)

// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		serviceReader := bytes.NewReader(jsonBytes)

// 		// 创建一个http请求，请求方法为POST，请求路径为"/api/v1/namespaces/:namespace/services"，请求体为一个json字符串。
// 		k8log.DebugLog("APIServer", "TestAddService: serviceReader = "+string(jsonBytes))
// 		req, err := http.NewRequest("POST", stringutil.Replace(config.ServiceURL, config.URL_PARAM_NAMESPACE_PART, service.Metadata.Namespace), serviceReader)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		req.Header.Set("Content-Type", "application/json")

// 		// 创建响应写入器
// 		w := httptest.NewRecorder()

// 		// 将请求和响应写入gin.Context
// 		c, _ := gin.CreateTestContext(w)
// 		c.Request = req

// 		// 执行处理函数
// 		r.HandleContext(c)

// 		// 获取响应结果
// 		resp := w.Result()

// 		if resp.StatusCode != http.StatusCreated {
// 			t.Errorf("expected status %v but got %v", http.StatusOK, resp.StatusCode)
// 		}
// 	}
// }
