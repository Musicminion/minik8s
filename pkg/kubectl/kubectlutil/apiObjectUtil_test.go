package kubectlutil

import (
	"miniK8s/util/file"
	"testing"
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
