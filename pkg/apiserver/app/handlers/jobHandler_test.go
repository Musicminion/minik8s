package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"miniK8s/pkg/apiObject"
	etcdclient "miniK8s/pkg/apiserver/app/etcdclient"
	"miniK8s/pkg/config"
	"miniK8s/util/stringutil"
	"miniK8s/util/zip"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

const ifClear = true

func TestMain(m *testing.M) {
	if ifClear {
		// 删除 ./testFile/zipFile/ 目录下的所有文件
		os.RemoveAll("./testFile/zipFile/outPut/")
		os.Mkdir("./testFile/zipFile/outPut/", os.ModePerm)

		// 清空etcd 中的数据
		res := etcdclient.EtcdStore.DelAll()

		if res != nil {
			panic(res)
		}
	}
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

	err := zip.CompressToZip("./testFile/zipFile/zipFolder", "./testFile/zipFile/outPut/test-job.zip")

	if err != nil {
		t.Fatal(err)
	}

	res, err := zip.ComvertZipToBytes("./testFile/zipFile/outPut/test-job.zip")

	if err != nil {
		t.Fatal(err)
	}

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

	// t.Logf("resp: %v", resp.Body)

	// 读取response body
	var result map[string]interface{}

	err = json.NewDecoder(resp.Body).Decode(&result)

	if err != nil {
		t.Fatal(err)
	}

	data, ok := result["data"]

	if !ok {
		t.Fatal("data not found")
	}

	dataStr := fmt.Sprint(data)

	jobfile := &apiObject.JobFile{}
	err = json.Unmarshal([]byte(dataStr), jobfile)

	if err != nil {
		t.Fatal(err)
	}

	if jobfile.Metadata.Name != "job1" {
		t.Errorf("expected jobfile name %s, got %s", "job1", jobfile.Metadata.Name)
	}

	err = zip.ConvertBytesToZip(jobfile.UserUploadFile, "./testFile/zipFile/outPut/test-job-resp.zip")

	if err != nil {
		t.Fatal(err)
	}

	res := zip.DecompressZip("./testFile/zipFile/outPut/test-job-resp.zip", "./testFile/zipFile/outPut/test-job-resp/")

	if res != nil {
		t.Fatal(res)
	}
}

func TestAddJob(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	// 创建一个新的gin引擎，并注册AddPod处理函数。
	r := gin.New()
	// 关闭gin的日志输出
	r.Use(gin.LoggerWithWriter(io.Discard))
	// 设置gin为生产模式
	gin.SetMode(gin.ReleaseMode)
	// 通过调用gin引擎的ServeHTTP方法，可以模拟一个http请求，从而测试AddPod方法。

	r.POST(config.JobsURL, AddJob)

	job := &apiObject.Job{
		Basic: apiObject.Basic{
			APIVersion: "v1",
			Kind:       "Job",
			Metadata: apiObject.Metadata{
				Name:      "job1",
				Namespace: "default",
			},
		},
		Spec: apiObject.JobSpec{
			JobPartition:    "dgx2",
			NTasks:          1,
			NTasksPerNode:   6,
			SubmitDirectory: "./testFile/zipFile/zipFolder",
			CompileCommands: []string{"ls"},
			RunCommands:     []string{"ls", "echo hello", "pwd", "echo 123"},
			OutputFile:      "test-out",
			ErrorFile:       "test-error",
			JobUsername:     os.Getenv("GPU_SSH_USERNAME"),
			JobPassword:     os.Getenv("GPU_SSH_PASSWORD"),
			GPUNums:         1,
		},
	}

	// 读取的内容转化为json
	jsonBytes, err := json.Marshal(job)

	if err != nil {
		t.Fatal(err)
	}

	jobReader := bytes.NewReader(jsonBytes)

	url := stringutil.Replace(config.JobsURL, config.URL_PARAM_NAMESPACE_PART, "default")

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

func TestClear(t *testing.T) {

	if ifClear {
		// 清空etcd 中的数据
		res := etcdclient.EtcdStore.DelAll()

		if res != nil {
			panic(res)
		}

		// 删除 ./testFile/zipFile/ 目录下的所有文件
		os.RemoveAll("./testFile/zipFile/outPut/")
		os.Mkdir("./testFile/zipFile/outPut/", os.ModePerm)
	}

}
