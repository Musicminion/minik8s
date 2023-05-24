package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"miniK8s/pkg/apiObject"
	etcdclient "miniK8s/pkg/apiserver/app/etcdclient"
	"miniK8s/pkg/config"
	"miniK8s/util/stringutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

var testDns = apiObject.Dns{
	Basic: apiObject.Basic{
		APIVersion: "v1",
		Kind:       "Dns",
		Metadata: apiObject.Metadata{
			Name:      "testDns",
			Namespace: "default",
		},
	},
	Spec: apiObject.DnsSpec{
		Host: "example.com",
		Paths: []apiObject.Path{
			{
				SubPath: "/api/v1",
				SvcName: "example-service1",
				SvcPort: "80",
				SvcIp:   "",
			},
		},
	},
}

var testService = apiObject.ServiceStore{
	Basic: apiObject.Basic{
		APIVersion: "v1",
		Kind:       "Service",
		Metadata: apiObject.Metadata{
			Name:      "example-service1",
			Namespace: "default",
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
				Name:       "testPort",
			},
		},
	},
}


func TestAddDns(t *testing.T) {
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
	r.POST(config.DnsURL, AddDns)
	r.POST(config.ServiceURL, AddService)

	// 通过调用gin引擎的ServeHTTP方法，可以模拟一个http请求，从而测试AddDns方法。
	
	// 创建svc
	svcJsonBytes, err := json.Marshal(testService)
	if err != nil {
		t.Fatal(err)
	}
	svcReader := bytes.NewReader(svcJsonBytes)
	svcURL := config.ServiceURL
	svcURL = stringutil.Replace(svcURL, config.URL_PARAM_NAMESPACE_PART, testService.Metadata.Namespace)
	req, err := http.NewRequest(http.MethodPost, svcURL, svcReader)
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
	time.Sleep(1 * time.Second)

	// 读取的内容转化为json
	dnsJsonBytes, err := json.Marshal(testDns)
	if err != nil {
		t.Fatal(err)
	}
	dnsReader := bytes.NewReader(dnsJsonBytes)

	URL := config.DnsURL
	URL = stringutil.Replace(URL, config.URL_PARAM_NAMESPACE_PART, testDns.Metadata.Namespace)

	req, err = http.NewRequest(http.MethodPost, URL, dnsReader)

	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	// 创建响应写入器
	w = httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp = w.Result()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Status code error: %d", resp.StatusCode)
	}
}

func TestGetDns(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	// 创建一个新的gin引擎，并注册GetDns处理函数。
	r := gin.New()
	// 关闭gin的日志输出
	r.Use(gin.LoggerWithWriter(io.Discard))
	// 设置gin为生产模式
	gin.SetMode(gin.ReleaseMode)
	// 通过调用gin引擎的ServeHTTP方法，可以模拟一个http请求，从而测试GetDns方法。
	r.GET(config.DnsSpecURL, GetDns)

	// 创建一个http请求，请求方法为GET，请求路径为"/api/v1/namespaces/:namespace/dns"。
	URL := stringutil.Replace(config.DnsSpecURL, config.URL_PARAM_NAMESPACE_PART, testDns.Metadata.Namespace)
	URL = stringutil.Replace(URL, config.URL_PARAM_NAME_PART, testDns.Metadata.Name)
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	// 创建响应写入器
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Status code error: %d", resp.StatusCode)
	}
}

func TestDeleteDns(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	// 创建一个新的gin引擎，并注册DeleteDns处理函数。
	r := gin.New()
	// 关闭gin的日志输出
	r.Use(gin.LoggerWithWriter(io.Discard))
	// 设置gin为生产模式
	gin.SetMode(gin.ReleaseMode)
	// 通过调用gin引擎的ServeHTTP方法，可以模拟一个http请求，从而测试DeleteDns方法。
	r.DELETE(config.DnsSpecURL, DeleteDns)

	// 创建一个http请求，请求方法为DELETE，请求路径为"/api/v1/namespaces/:namespace/dns"。
	URL := stringutil.Replace(config.DnsSpecURL, config.URL_PARAM_NAMESPACE_PART, testDns.Metadata.Namespace)
	URL = stringutil.Replace(URL, config.URL_PARAM_NAME_PART, testDns.Metadata.Name)
	req, err := http.NewRequest(http.MethodDelete, URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	// 创建响应写入器
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Status code error: %d", resp.StatusCode)
	}
}
