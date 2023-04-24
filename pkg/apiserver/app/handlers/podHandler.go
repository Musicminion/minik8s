package handlers

import (
	"encoding/json"
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/k8log"

	"miniK8s/util/uuid"

	"github.com/gin-gonic/gin"
)

// 获取单个Pod的信息
// "/api/v1/namespaces/:namespace/pods"
func GetPod(c *gin.Context) {
	// "/api/v1/namespaces/:namespace/pods/:name"
	// 解析里面的参数
	namespace := c.Param("namespace")
	name := c.Param("name")
	if namespace == "" || name == "" {
		c.JSON(400, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}
	// 从etcd中获取
	// ETCD里面的路径是 /registry/pods/<namespace>/<pod-name>
	logStr := fmt.Sprintf("GetPod: namespace = %s, name = %s", namespace, name)
	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf("/registry/pods/%s/%s", namespace, name)
	res, err := EtcdStore.Get(key)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "get pod failed " + err.Error(),
		})
		return
	}

	if len(res) == 0 {
		c.JSON(404, gin.H{
			"error": "get pod err, not find pod",
		})
		return
	}

	// 处理Res，如果有多个返回的，报错
	if len(res) != 1 {
		c.JSON(500, gin.H{
			"error": "get pod err, find more than one pod",
		})
		return
	}
	// 遍历res，返回对应的Node信息
	targetPod := res[0].Value
	c.JSON(200, gin.H{
		"data": targetPod,
	})
}

// 获取所有的Pod的信息
// API "/api/v1/namespaces/:namespace/pods"
func GetPods(c *gin.Context) {
	namespaces := c.Param("namespace")
	if namespaces == "" {
		c.JSON(400, gin.H{
			"error": "namespace is empty",
		})
		return
	}
	// 从etcd中获取
	// ETCD里面的路径是 /registry/pods/<namespace>/
	logStr := fmt.Sprintf("GetPods: namespace = %s", namespaces)
	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf("/registry/pods/%s/", namespaces)
	res, err := EtcdStore.PrefixGet(key)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "get pods failed " + err.Error(),
		})
		return
	}

	if len(res) == 0 {
		c.JSON(404, gin.H{
			"error": "get pods err, not find pods",
		})
		return
	}

	// 遍历res，返回对应的Node信息
	targetPods := make([]string, 0)
	for _, pod := range res {
		targetPods = append(targetPods, pod.Value)
	}
	c.JSON(200, gin.H{
		"data": targetPods,
	})
}

// POST "/api/v1/namespaces/:namespace/pods"
func AddPod(c *gin.Context) {
	// log
	k8log.InfoLog("APIServer", "AddPod")

	// POST /api/v1/namespaces/:namespace/pods
	// 解析里面的参数
	namespace := c.Param("namespace")
	if namespace == "" {
		c.JSON(400, gin.H{
			"error": "namespace is empty",
		})
		return
	}

	// 从body中获取pod的信息
	var pod apiObject.Pod
	if err := c.ShouldBind(&pod); err != nil {
		c.JSON(500, gin.H{
			"error": "parser pod failed " + err.Error(),
		})

		k8log.ErrorLog("APIServer", "AddPod: parser pod failed "+err.Error())
		return
	}

	// 检查名字是否为空
	newPodName := pod.Metadata.Name
	if newPodName == "" {
		c.JSON(400, gin.H{
			"error": "pod name is empty",
		})
		return
	}

	// 检查name是否重复
	key := fmt.Sprintf("/registry/pods/%s/%s", namespace, newPodName)
	res, err := EtcdStore.Get(key)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "get pod failed " + err.Error(),
		})
		return
	}

	if len(res) != 0 {
		c.JSON(400, gin.H{
			"error": "pod name has exist",
		})
		return
	}

	// 给Pod设置UUID，用于后面的调度
	// 哪怕用户自己设置了UUID，也会被覆盖
	pod.Metadata.UUID = uuid.NewUUID()

	// 把Pod转化为PodStore
	podStore := pod.ToStore()

	// 把PodStore转化为json
	podStoreJson, err := json.Marshal(podStore)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "parser pod to json failed " + err.Error(),
		})
		return
	}

	// 将pod存储到etcd中
	// 持久化
	key = fmt.Sprintf("/registry/pods/%s/%s", namespace, newPodName)

	// 将pod存储到etcd中
	err = EtcdStore.Put(key, podStoreJson)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "put pod to etcd failed " + err.Error(),
		})
		return
	}

	// 返回
	c.JSON(201, gin.H{
		"message": "create pod success",
	})

	/*
		后面需要发送请求给调度器，让调度器进行调度到节点上面
	*/
}

// "/api/v1/namespaces/:namespace/pods/:name"
func DeletePod(c *gin.Context) {
	// log
	k8log.InfoLog("APIServer", "DeletePod")

	// 解析参数
	namespace := c.Param("namespace")
	name := c.Param("name")

	// 检查参数是否为空
	if namespace == "" || name == "" {
		c.JSON(400, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}

	logStr := fmt.Sprintf("DeletePod: namespace = %s, name = %s", namespace, name)
	k8log.InfoLog("APIServer", logStr)

	err := EtcdStore.Del(fmt.Sprintf("/registry/pods/%s/%s", namespace, name))

	if err != nil {
		c.JSON(500, gin.H{
			"error": "delete pod failed " + err.Error(),
		})
		return
	}

	c.JSON(204, gin.H{
		"message": "delete pod success",
	})

}

func UpdatePod(c *gin.Context) {
	// TODO: 更新pod
}

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io"

// 	"github.com/gin-gonic/gin"

// 	"miniK8s/pkg/apiObject"
// 	"miniK8s/pkg/apiserver/config"
// )

// func (s *apiServer) AddPod(c *gin.Context) {
// 	body, _ := io.ReadAll(c.Request.Body)
// 	// 对pod进行赋值
// 	pod := &apiObject.Pod{}
// 	err := json.Unmarshal(body, pod)
// 	if err != nil {
// 		fmt.Println("[AddPod] unmarshall pod fail")
// 		return
// 	}

// 	body, _ = json.Marshal(pod)
// 	// TODO: 先判断pod是否已存在

// 	// 持久化
// 	err = s.etcdStore.Put("to filled"+"/"+pod.Name, body)
// 	if err != nil {
// 		fmt.Println("[AddPod] etcd failed to put")
// 	}
// }

// // A naive delete method
// func (s *apiServer) DeletePod(c *gin.Context) {
// 	podName := c.Param(config.ResourceName)
// 	key := "" + podName + "" // to be defined

// 	// 实际上，仅仅对pod进行删除是不够的
// 	if err := s.etcdStore.Del(key); err == nil {
// 		fmt.Println("[DelPod] etcd failed to delete")
// 	} else {
// 		fmt.Printf("Delete pod %s successfully\n", podName)
// 	}
// }
