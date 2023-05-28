package apiObject

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestServiceYaml(t *testing.T) {

	file, err := os.Open("./testFile/yamlFile/Service.yaml")
	if err != nil {
		t.Fatal(err)
	}
	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	// t.Log(string(content))

	// 将文件内容转换为Node对象
	service := &Service{}
	err = yaml.Unmarshal(content, service)

	if err != nil {
		t.Fatal(err)
	}
	// 比较转换后的Node对象与预期的Node对象是否相同
	// 输出转换后的Node对象
	t.Log(service.GetAPIVersion())
	t.Log(service.GetType())
	t.Log(service.GetObjectKind())
	t.Log(service.GetPorts())
	t.Log(service.GetObjectName())
}

func TestServiceJson(t *testing.T) {
	// 打开文件
	file, err := os.Open("./testFile/jsonFile/Service.json")
	if err != nil {
		t.Fatal(err)
	}
	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	// t.Log(string(content))

	// 将文件内容转换为Node对象
	service := &Service{}
	err = json.Unmarshal(content, service)

	if err != nil {
		t.Fatal(err)
	}
	// 比较转换后的Node对象与预期的Node对象是否相同
	// 输出转换后的Node对象
	t.Log(service.GetAPIVersion())
	t.Log(service.GetType())
	t.Log(service.GetObjectKind())
	t.Log(service.GetPorts())
	t.Log(service.GetObjectName())
}
