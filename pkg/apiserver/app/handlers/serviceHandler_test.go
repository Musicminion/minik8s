package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"miniK8s/pkg/apiObject"
	etcdclient "miniK8s/pkg/apiserver/app/etcdclient"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	"miniK8s/util/stringutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

func TestAddService(t *testing.T) {
	// 清空etcd
	etcdclient.EtcdStore.PrefixDel("/")

	gin.SetMode(gin.ReleaseMode)
	// 创建一个新的gin引擎，并注册AddService处理函数。
	r := gin.New()
	// 关闭gin的日志输出
	r.Use(gin.LoggerWithWriter(io.Discard))
	// 设置gin为生产模式
	gin.SetMode(gin.ReleaseMode)
	// 通过调用gin引擎的ServeHTTP方法，可以模拟一个http请求，从而测试AddService方法。
	r.POST(config.ServiceURL, AddService)

	// 读取文件"./testFile/yamlFile/Service-i.yaml"，将文件内容作为请求体。
	// 打开文件
	for i := 1; i <= 2; i++ {
		path := "./testFile/yamlFile/Service-" + fmt.Sprint(i) + ".yaml"
		file, err := os.Open(path)

		if err != nil {
			t.Fatal(err)
		}
		// 读取文件内容
		content, err := io.ReadAll(file)
		if err != nil {
			t.Fatal(err)
		}

		// 将文件内容转换为Service对象
		// 通过调用gin引擎的ServeHTTP方法，可以模拟一个http请求，从而测试AddService方法。
		service := &apiObject.ServiceStore{}
		err = yaml.Unmarshal(content, service)

		if err != nil {
			t.Fatal(err)
		}
		// 读取的内容转化为json

		jsonBytes, err := json.Marshal(service)

		if err != nil {
			t.Fatal(err)
		}
		serviceReader := bytes.NewReader(jsonBytes)

		// 创建一个http请求，请求方法为POST，请求路径为"/api/v1/namespaces/:namespace/services"，请求体为一个json字符串。
		k8log.DebugLog("APIServer", "TestAddService: serviceReader = "+string(jsonBytes))
		req, err := http.NewRequest("POST", stringutil.Replace(config.ServiceURL, config.URL_PARAM_NAMESPACE_PART, service.Metadata.Namespace), serviceReader)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		// 创建响应写入器
		w := httptest.NewRecorder()

		// 将请求和响应写入gin.Context
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// 执行处理函数
		r.HandleContext(c)

		// 获取响应结果
		resp := w.Result()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("expected status %v but got %v", http.StatusOK, resp.StatusCode)
		}
	}
}

func TestGetServices(t *testing.T) {
	// 创建一个新的gin引擎，并注册GetService处理函数。
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output: io.Discard, // 将输出重定向到 ioutil.Discard，即丢弃
	}))
	r.GET(config.ServiceSpecURL, GetServices)

	for i := 1; i <= 2; i++ {
		// 创建一个http请求，请求方法为GET，请求路径为"/api/v1/services"。

		uri := config.ServiceURL + "/service-example" + fmt.Sprint(i)
		req, err := http.NewRequest("GET", uri, nil)
		req.Header.Set("Content-Type", "application/json")
		if err != nil {

			t.Fatal(err)
		}

		// 创建响应写入器
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		// 检查响应状态码和响应体
		if w.Code != http.StatusOK {
			t.Errorf("expected status %v but got %v", http.StatusOK, w.Code)
		}

	}
}

func TestDeleteService(t *testing.T) {
	// // 创建一个新的gin引擎，并注册DeleteService处理函数。
	// r := gin.Default()
	// // 把r的输出重定向到null
	// r.Use(gin.LoggerWithWriter(io.Discard))
	// 创建一个新的gin引擎，并注册GetService处理函数。
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output: io.Discard, // 将输出重定向到 ioutil.Discard，即丢弃
	}))
	r.DELETE(config.ServiceSpecURL, DeleteService)

	for i := 1; i <= 2; i++ {
		// 创建一个http请求，请求方法为GET，请求路径为"/api/v1/services"。
		uri := config.ServiceSpecURL
		uri = stringutil.Replace(uri, config.URL_PARAM_NAMESPACE_PART, "default")
		uri = stringutil.Replace(uri, config.URL_PARAM_NAME_PART, "service-example"+strconv.Itoa(i))
		req, err := http.NewRequest("DELETE", uri, nil)
		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			t.Fatal(err)
		}

		// 创建响应写入器
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		// 检查响应状态码和响应体
		if w.Code != http.StatusNoContent {
			t.Errorf("expected status %v but got %v", http.StatusNoContent, w.Code)
		}
		t.Log(w.Body.String())
	}

}

// func TestAddServiceWithEndpoints(t *testing.T) {
// 	// 清空etcd
// 	etcdclient.EtcdStore.PrefixDel("/")

// 	gin.SetMode(gin.ReleaseMode)
// 	// 创建一个新的gin引擎，并注册AddService处理函数。
// 	r := gin.New()
// 	// 关闭gin的日志输出
// 	r.Use(gin.LoggerWithWriter(io.Discard))
// 	// 设置gin为生产模式
// 	gin.SetMode(gin.ReleaseMode)
// 	// 通过调用gin引擎的ServeHTTP方法，可以模拟一个http请求，从而测试AddService方法。
// 	r.POST(config.ServiceURL, AddService)

// 	// 确保当前etcd中有含有label的Pod
// 	testutil.AddPodToEtcd()

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
// 			t.Errorf("expected status %v but got %v", http.StatusCreated, resp.StatusCode)
// 		}
// 		// 清空etcd
// 		etcdclient.EtcdStore.PrefixDel("/")
// 	}
// }
