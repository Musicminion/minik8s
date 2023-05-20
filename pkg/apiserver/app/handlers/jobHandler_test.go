package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/util/stringutil"
	"miniK8s/util/zip"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	// 删除 ./testFile/zipFile/ 目录下的所有文件
	os.RemoveAll("./testFile/zipFile/outPut/")
	os.Mkdir("./testFile/zipFile/outPut/", os.ModePerm)
	m.Run()
}

func TestAddJobFile(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	// 创建一个新的gin引擎，并注册AddPod处理函数。
	r := gin.New()
	// 关闭gin的日志输出
	r.Use(gin.LoggerWithWriter(io.Discard))
	// 设置gin为生产模式
	gin.SetMode(gin.ReleaseMode)
	// 通过调用gin引擎的ServeHTTP方法，可以模拟一个http请求，从而测试AddPod方法。

	r.POST(config.JobFileURL, AddJobFile)

	jobfile := &apiObject.JobFile{
		Basic: apiObject.Basic{
			APIVersion: "v1",
			Kind:       "Job",
			Metadata: apiObject.Metadata{
				Name:      "job1",
				Namespace: "default",
			},
		},
	}

	zip.CompressToZip("./testFile/zipFile/zipFolder/", "./testFile/zipFile/outPut/test-job.zip")

	res, err := zip.ComvertZipToBytes("./testFile/zipFile/outPut/test-job.zip")

	if err != nil {
		t.Fatal(err)
	}

	jobfile.UserUploadFile = res

	// 读取的内容转化为json
	jsonBytes, err := json.Marshal(jobfile)

	if err != nil {
		t.Fatal(err)
	}

	jobReader := bytes.NewReader(jsonBytes)

	url := stringutil.Replace(config.JobFileURL, config.URL_PARAM_NAMESPACE_PART, "default")

	req, err := http.NewRequest(http.MethodPost, url, jobReader)

	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 创建响应写入器
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Status code error: %d", resp.StatusCode)
	}
}

func TestGetJobFile(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	// 创建一个新的gin引擎，并注册AddPod处理函数。
	r := gin.New()
	// 关闭gin的日志输出
	r.Use(gin.LoggerWithWriter(io.Discard))
	// 设置gin为生产模式
	gin.SetMode(gin.ReleaseMode)
	// 通过调用gin引擎的ServeHTTP方法，可以模拟一个http请求，从而测试AddPod方法。

	r.GET(config.JobFileSpecURL, GetJobFile)

	// url := stringutil.Replace(config.JobFileSpecURL, config.URL_PARAM_NAMESPACE_PART, "default")
	// url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, "job1")

	url := "/apis/v1/namespaces/default/jobfiles/job1"

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	// 创建响应写入器
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
