package handlers

import (
	"fmt"
	etcdclient "miniK8s/pkg/apiserver/app/etcdclient"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	"miniK8s/util/stringutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

// "/apis/v1/namespaces/:namespace/hpa/:name"
func GetHPA(c *gin.Context) {
	namespace := c.Param(config.URL_PARAM_NAMESPACE)
	name := c.Param(config.URL_PARAM_NAME)
	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}

	logStr := fmt.Sprintf("GetHPA: namespace=%s, name=%s", namespace, name)

	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf(serverconfig.EtcdHpaPath+"%s/%s", namespace, name)
	res, err := etcdclient.EtcdStore.Get(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "GetHPA: " + err.Error(),
		})
		return
	}

	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "GetHPA: not found",
		})
		return
	}

	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "GetHPA: more than one result",
		})
		return
	}

	targetHPA := res[0].Value

	c.JSON(http.StatusOK, gin.H{
		"data": targetHPA,
	})
}

// GET 获取某个名字空间下的所有HPA
// "/apis/v1/namespaces/:namespace/hpa"
func GetHPAs(c *gin.Context) {
	namespace := c.Param(config.URL_PARAM_NAMESPACE)
	if namespace == "" {
		namespace = config.DefaultNamespace
	}

	logStr := fmt.Sprintf("GetHPA: namespace=%s", namespace)
	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf(serverconfig.EtcdHpaPath+"%s/", namespace)

	res, err := etcdclient.EtcdStore.PrefixGet(key)

	if err != nil {
		k8log.ErrorLog("APIServer", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "GetHPA: " + err.Error(),
		})
		return
	}

	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "GetHPA: not found",
		})
		return
	}

	targetHPAs := make([]string, 0)
	for _, hpa := range res {
		targetHPAs = append(targetHPAs, string(hpa.Value))
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stringutil.StringSliceToJsonArray(targetHPAs),
	})
}

// POST 创建一个HPA
// "/apis/v1/namespaces/:namespace/hpa"
func AddHPA(c *gin.Context) {

}

// PUT 更新一个HPA
// "/apis/v1/namespaces/:namespace/hpa/:name"
func UpdateHPA(c *gin.Context) {

}

// DELETE 删除一个HPA
// "/apis/v1/namespaces/:namespace/hpa/:name"
func DeleteHPA(c *gin.Context) {

}

// GET 获取全局的HPA
// "/apis/v1/hpa"
func GetGlobalHPAs(c *gin.Context) {

}
