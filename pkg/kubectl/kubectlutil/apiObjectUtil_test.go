package kubectlutil

import (
	"miniK8s/util/file"
	"testing"
)

func TestGetAPIObjectTypeFromYamlFile(t *testing.T) {
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
