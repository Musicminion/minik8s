package handlers

import (
	"encoding/json"
	"fmt"
	"miniK8s/pkg/apiObject"
	etcdclient "miniK8s/pkg/apiserver/app/etcdclient"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	"miniK8s/util/stringutil"
	"miniK8s/util/uuid"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetFunctions
// "/apis/v1/namespaces/:namespace/functions"
func GetFunctions(c *gin.Context) {
	namespace := c.Param(config.URL_PARAM_NAMESPACE)
	if namespace == "" {
		namespace = config.DefaultNamespace
	}

	logStr := fmt.Sprintf("GetFunctions: namespace=%s", namespace)
	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf(serverconfig.EtcdFunctionPath+"%s", namespace)

	res, err := etcdclient.EtcdStore.PrefixGet(key)

	if err != nil {
		k8log.ErrorLog("APIServer", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "GetFunctions: " + err.Error(),
		})
		return
	}

	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "GetFunctions: not found",
		})
		return
	}

	targetFunc := make([]string, 0)

	for _, fun := range res {
		targetFunc = append(targetFunc, fun.Value)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stringutil.StringSliceToJsonArray(targetFunc),
	})
}

// GetFunction
// "/apis/v1/namespaces/:namespace/functions/:name"
func GetFunction(c *gin.Context) {
	namespace := c.Param(config.URL_PARAM_NAMESPACE)
	name := c.Param(config.URL_PARAM_NAME)
	// 检查参数
	if namespace == "" {
		namespace = config.DefaultNamespace
	}

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "name is empty",
		})
		k8log.ErrorLog("APIServer", "GetFunction: name is empty")
		return
	}

	logStr := fmt.Sprintf("GetFunction: namespace=%s, name=%s", namespace, name)
	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf(serverconfig.EtcdFunctionPath+"%s/%s", namespace, name)
	res, err := etcdclient.EtcdStore.Get(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "GetFunction: " + err.Error(),
		})
		return
	}

	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "GetFunction: not found",
		})
		return
	}

	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "GetFunction: " + err.Error(),
		})
		return
	}

	targetFunc := res[0].Value

	c.JSON(http.StatusOK, gin.H{
		"data": targetFunc,
	})
}

// AddFunction
func AddFunction(c *gin.Context) {
	// log
	k8log.InfoLog("APIServer", "AddFunction")

	// 从请求中获取参数
	var newfunction apiObject.Function
	if err := c.ShouldBindJSON(&newfunction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "AddFunction: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "AddFunction: "+err.Error())
		return
	}

	newFuncName := newfunction.Metadata.Name

	if newFuncName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "AddFunction: name is empty",
		})
		k8log.ErrorLog("APIServer", "AddFunction: name is empty")
		return
	}

	if newfunction.Metadata.Namespace == "" {
		newfunction.Metadata.Namespace = config.DefaultNamespace
	}

	// 检查是否存在
	key := fmt.Sprintf(serverconfig.EtcdFunctionPath+"%s/%s", newfunction.Metadata.Namespace, newFuncName)

	res, err := etcdclient.EtcdStore.Get(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "AddFunction: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "AddFunction: "+err.Error())
		return
	}

	if len(res) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "AddFunction: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "AddFunction: "+err.Error())
		return
	}

	newfunction.Metadata.UUID = uuid.NewUUID()

	newfunctionJson, err := json.Marshal(newfunction)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "AddFunction: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "AddFunction: "+err.Error())
		return
	}

	key = fmt.Sprintf(serverconfig.EtcdFunctionPath+"%s/%s", newfunction.Metadata.Namespace, newFuncName)

	err = etcdclient.EtcdStore.Put(key, newfunctionJson)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "AddFunction: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "AddFunction: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "AddFunction: success",
	})
	/*
		后面如果要做什么再加
	*/
}

// UpdateFunction
func UpdateFunction(c *gin.Context) {

}

// DeleteFunction
func DeleteFunction(c *gin.Context) {
	k8log.InfoLog("APIServer", "DeleteFunction")

	namespace := c.Param(config.URL_PARAM_NAMESPACE)
	name := c.Param(config.URL_PARAM_NAME)

	if namespace == "" {
		namespace = config.DefaultNamespace
	}

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "DeleteFunction: name is empty",
		})
		k8log.ErrorLog("APIServer", "DeleteFunction: name is empty")
		return
	}

	key := fmt.Sprintf(serverconfig.EtcdFunctionPath+"%s/%s", namespace, name)

	k8log.InfoLog("APIServer", "DeleteFunction: key="+key)

	err := etcdclient.EtcdStore.Del(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "DeleteFunction: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "DeleteFunction: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "DeleteFunction: success",
	})

}

// GetGlobalFunctions
func GetGlobalFunctions(c *gin.Context) {
	k8log.InfoLog("APIServer", "GetGlobalFunctions")

	key := fmt.Sprintf(serverconfig.EtcdFunctionPath)

	res, err := etcdclient.EtcdStore.PrefixGet(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "GetGlobalFunctions: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "GetGlobalFunctions: "+err.Error())
		return
	}

	targetFunc := make([]string, 0)

	for _, fun := range res {
		targetFunc = append(targetFunc, fun.Value)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stringutil.StringSliceToJsonArray(targetFunc),
	})

}
