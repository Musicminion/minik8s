package apiObject

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYaml(t *testing.T) {
	// 打开文件
	file, err := os.Open("./testFile/yamlFile/Node.yaml")
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
	node := &Node{}
	err = yaml.Unmarshal(content, node)

	if err != nil {
		t.Fatal(err)
	}
	// 比较转换后的Node对象与预期的Node对象是否相同
	// 输出转换后的Node对象
	t.Log(node.GetAPIVersion())
	t.Log(node.GetObjectKind())
	t.Log(node.GetAnnotations())
	t.Log(node.GetLabels())
	t.Log(node.GetUUID())
	t.Log(node.GetObjectName())
	t.Log(node.GetIP())
}

func TestJson(t *testing.T) {
	// 打开文件
	file, err := os.Open("./testFile/jsonFile/Node.json")
	if err != nil {
		t.Fatal(err)
	}
	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	// t.Log(string(content))

	// 将文件内容转换为Node的JSON对象
	node := &Node{}
	err = json.Unmarshal(content, node)

	if err != nil {
		t.Fatal(err)
	}
	// 比较转换后的Node对象与预期的Node对象是否相同
	// 输出转换后的Node对象
	t.Log(node.GetAPIVersion())
	t.Log(node.GetObjectKind())
	t.Log(node.GetAnnotations())
	t.Log(node.GetLabels())
	t.Log(node.GetUUID())
	t.Log(node.GetObjectName())
	t.Log(node.GetIP())
}
