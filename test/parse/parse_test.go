package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	// "path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	// "github.com/ugorji/go/codec"
	"gopkg.in/yaml.v3"

	"miniK8s/pkg/apiObject"
)

func TestMain(t *testing.T) {
	// 读取yaml文件
	podYamlPath := "/miniK8s/test/parse/test.yaml"
	fd, err := os.Open(podYamlPath)
	assert.Nil(t, err)

	content, err := ioutil.ReadAll(fd)
	// t.Error(content)
	assert.Nil(t, err)

	// 把Yaml变成pod
	pod := &apiObject.Pod{}
	err = yaml.Unmarshal(content, &pod)
	// t.Error(pod)
	assert.Nil(t, err)

	// send apiobject to apiserver

	cli := http.Client{}
	b, _ := json.Marshal(pod)
	req, err := http.NewRequest(http.MethodGet, "localhost:8080/get/666", bytes.NewReader(b))
	assert.Nil(t, err)
	t.Error(req)
	resp, err := cli.Do(req)
	assert.Nil(t, err)

	respBody, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	fmt.Println(string(respBody))

}
