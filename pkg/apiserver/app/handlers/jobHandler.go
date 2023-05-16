package handlers

import (
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
