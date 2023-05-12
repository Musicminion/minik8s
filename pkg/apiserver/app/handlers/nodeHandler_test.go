// 测试NodeHandler的方法
package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

func TestAddNode(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	// 创建一个新的gin引擎，并注册AddNode处理函数。
	r := gin.New()
	// 关闭gin的日志输出
	r.Use(gin.LoggerWithWriter(io.Discard))
	// 设置gin为生产模式
	gin.SetMode(gin.ReleaseMode)
	// 通过调用gin引擎的ServeHTTP方法，可以模拟一个http请求，从而测试AddNode方法。
	r.POST(config.NodesURL, AddNode)

	// 读取文件"./testFile/yamlFile/Node-i.yaml"，将文件内容作为请求体。
	// 打开文件

	for i := 1; i <= 2; i++ {
		path := "./testFile/yamlFile/Node-" + fmt.Sprint(i) + ".yaml"
		file, err := os.Open(path)

		if err != nil {
			t.Fatal(err)
		}
		// 读取文件内容
		content, err := io.ReadAll(file)
		if err != nil {
			t.Fatal(err)
		}

		// 将文件内容转换为Node对象
		// 通过调用gin引擎的ServeHTTP方法，可以模拟一个http请求，从而测试AddNode方法。
		node := &apiObject.NodeStore{}
		err = yaml.Unmarshal(content, node)

		if err != nil {
			t.Fatal(err)
		}
		// 读取的内容转化为json

		jsonBytes, err := json.Marshal(node)

		if err != nil {
			t.Fatal(err)
		}
		nodeReader := bytes.NewReader(jsonBytes)

		// 创建一个http请求，请求方法为POST，请求路径为"/api/v1/nodes"，请求体为一个json字符串。
		req, err := http.NewRequest("POST", config.NodesURL, nodeReader)
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

func TestGetNodes(t *testing.T) {
	// 创建一个新的gin引擎，并注册GetNode处理函数。
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output: io.Discard, // 将输出重定向到 ioutil.Discard，即丢弃
	}))
	r.GET(config.NodeSpecURL, GetNodes)

	for i := 1; i <= 2; i++ {
		// 创建一个http请求，请求方法为GET，请求路径为"/api/v1/nodes"。
		uri := config.NodesURL + "/testNode-" + fmt.Sprint(i)
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

func TestDeleteNode(t *testing.T) {
	// // 创建一个新的gin引擎，并注册DeleteNode处理函数。
	// r := gin.Default()
	// // 把r的输出重定向到null
	// r.Use(gin.LoggerWithWriter(io.Discard))
	// 创建一个新的gin引擎，并注册GetNode处理函数。
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output: io.Discard, // 将输出重定向到 ioutil.Discard，即丢弃
	}))
	r.DELETE(config.NodeSpecURL, DeleteNode)

	for i := 1; i <= 2; i++ {
		// 创建一个http请求，请求方法为GET，请求路径为"/api/v1/nodes"。
		uri := config.NodesURL + "/testNode" + fmt.Sprint(i)
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
			t.Errorf("expected status %v but got %v", http.StatusOK, w.Code)
		}
		t.Log(w.Body.String())
	}

}
