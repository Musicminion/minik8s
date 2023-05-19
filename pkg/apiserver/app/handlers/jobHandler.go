package handlers

import (
	"fmt"
	"miniK8s/pkg/apiserver/app/etcdclient"
	"net/http"

	"github.com/gin-gonic/gin"
)

// // Job相关操作的URL
// JobsURL = "/api/v1/namespaces/:namespace/job"
// // 某个特定Job的URL
// JobSpecURL = "/api/v1/namespaces/:namespace/job/:name"
// // 某个特定Job的status
// JobSpecStatusURL = "/api/v1/namespaces/:namespace/job/:name/status"

// 获取单个的Job// JobSpecURL = "/api/v1/namespaces/:namespace/job/:name"
func GetJob(c *gin.Context) {
	namespace := c.Param("name")
	name := c.Param("name")
	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}
	// 从etcd中获取
	// ETCD里面的路径是 /registry/jods/<namespace>/<jod-name>
	key := fmt.Sprintf("/registry/jods/%s/%s", namespace, name)
	res, err := etcdclient.EtcdStore.Get(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get pod failed " + err.Error(),
		})
		return
	}

	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "pod not found",
		})
		return
	}

	// 处理Res，如果有多个返回的，报错
	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get pod err, find more than one pod",
		})
		return
	}

	// 遍历res，返回对应的Node信息
	targetJob := res[0].Value
	c.JSON(200, gin.H{
		"data": targetJob,
	})
}

// 获取所有的Job
func GetJobs(c *gin.Context) {
	// name := c.Param("name")
	// //job, err := api.GetJob(jobID)

	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{
	// 		"error": err.Error(),
	// 	})
	// 	return
	// }
	// c.JSON(http.StatusOK, job)
}
