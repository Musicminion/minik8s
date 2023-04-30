package kubectlutil

import (
	"bytes"
	"encoding/json"
	"io"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/apiserver/app/handlers"
	"strings"

	// apiserver "miniK8s/pkg/apiserver/app/server"
	"miniK8s/pkg/config"
	"miniK8s/util/file"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	// "miniK8s/pkg/kubectl/kubectlutil"
)

func TestGetAPIObjectTypeFromPodYamlFile(t *testing.T) {
	// 读取文件
	content, err := file.ReadFile("./testFile/pod.yaml")
	if err != nil {
		t.Fatal(err)
	}
	// 把文件内容转换成API对象
	kind, err := GetAPIObjectTypeFromYamlFile(content)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(kind)
}


func TestGetAPIObjectTypeFromServiceYamlFile(t *testing.T) {
	// 读取文件
	content, err := file.ReadFile("./testFile/service.yaml")
	if err != nil {
		t.Fatal(err)
	}
	// 把文件内容转换成API对象
	kind, err := GetAPIObjectTypeFromYamlFile(content)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(kind)
}

func TestParseAPIObjectFromYamlfileContent(t *testing.T) {
	fileContent, err := file.ReadFile("./testFile/pod.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var service apiObject.Service
	err = ParseAPIObjectFromYamlfileContent(fileContent, &service)
	if err != nil{
		t.Fatal(err)
	}
}

func TestPostAPIObjectToServer(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output: io.Discard, // 将输出重定向到 ioutil.Discard，即丢弃
	}))
	r.POST(config.ServiceURL, handlers.AddService)

	fileContent, err := file.ReadFile("./testFile/service.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var service apiObject.Service
	err = ParseAPIObjectFromYamlfileContent(fileContent, &service)
	if err != nil{
		t.Fatal(err)
	}

	jsonBytes, _ := json.Marshal(service)

	URL := config.ServiceURL
	URL = strings.Replace(URL, config.URL_PARAM_NAMESPACE_PART, service.Metadata.Namespace, -1)

	serviceHeader := bytes.NewReader(jsonBytes)
	req, err := http.NewRequest("POST", URL, serviceHeader)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	// err = PostAPIObjectToServer(URL, service)
	// if err != nil {
	// 	t.Fatal(err)
	// }
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

