package apiserver

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"

	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/apiserver/config"
)

func (s *apiServer) AddPod(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	// 对pod进行赋值
	pod := &apiObject.Pod{}
	err := json.Unmarshal(body, pod)
	if err != nil {
		fmt.Println("[AddPod] unmarshall pod fail")
		return
	}

	body, _ = json.Marshal(pod)
	// TODO: 先判断pod是否已存在

	// 持久化
	err = s.etcdStore.Put("to filled"+"/"+pod.Name, body)
	if err != nil {
		fmt.Println("[AddPod] etcd failed to put")
	}
}

// A naive delete method
func (s *apiServer) DeletePod(c *gin.Context) {
	podName := c.Param(config.ResourceName)
	key := "" + podName + "" // to be defined

	// 实际上，仅仅对pod进行删除是不够的
	if err := s.etcdStore.Del(key); err == nil {
		fmt.Println("[DelPod] etcd failed to delete")
	} else {
		fmt.Printf("Delete pod %s successfully\n", podName)
	}
}
