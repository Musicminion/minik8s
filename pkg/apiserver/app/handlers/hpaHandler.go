package handlers

import (
	"fmt"
	"miniK8s/pkg/apiObject"
	etcdclient "miniK8s/pkg/apiserver/app/etcdclient"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	"miniK8s/util/stringutil"
	"miniK8s/util/uuid"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
)

// "/apis/v1/namespaces/:namespace/hpa/:name"
func GetHPA(c *gin.Context) {
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
		k8log.ErrorLog("APIServer", "GetHPA: name is empty")
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
	// log
	k8log.InfoLog("APIServer", "AddHPA")

	// 从请求中获取HPA
	var hpa apiObject.HPA
	if err := c.ShouldBindJSON(&hpa); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "AddHPA: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", err.Error())
		return
	}

	newHPAName := hpa.Metadata.Name
	if newHPAName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "AddHPA: name is empty",
		})
		return
	}

	if hpa.Metadata.Namespace == "" {
		hpa.Metadata.Namespace = config.DefaultNamespace
	}

	// 检查是否已经存在
	key := fmt.Sprintf(serverconfig.EtcdHpaPath+"%s/%s", hpa.Metadata.Namespace, newHPAName)

	res, err := etcdclient.EtcdStore.Get(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "AddHPA: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", err.Error())
		return
	}

	if len(res) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "AddHPA: " + "already exists",
		})
		return
	}

	hpa.Metadata.UUID = uuid.NewUUID()

	// 把hpa转化为hpastore
	hpaStore := hpa.ToHPAStore()

	// 把hpaStore存入etcd
	hpaStoreJson, err := json.Marshal(hpaStore)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "AddHPA: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", err.Error())
		return
	}

	key = fmt.Sprintf(serverconfig.EtcdHpaPath+"%s/%s", hpa.Metadata.Namespace, newHPAName)

	err = etcdclient.EtcdStore.Put(key, hpaStoreJson)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "AddHPA: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "AddHPA: success",
	})

	/*
		后面如果要做什么再加
	*/
}

// PUT 更新一个HPA
// "/apis/v1/namespaces/:namespace/hpa/:name"
func UpdateHPA(c *gin.Context) {

}

// DELETE 删除一个HPA
// "/apis/v1/namespaces/:namespace/hpa/:name"
func DeleteHPA(c *gin.Context) {
	// log
	k8log.InfoLog("APIServer", "DeleteHPA")

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
		k8log.ErrorLog("APIServer", "DeletePod: name is empty")
		return
	}

	key := fmt.Sprintf(serverconfig.EtcdHpaPath+"%s/%s", namespace, name)

	k8log.InfoLog("APIServer", "DeleteHPA: key="+key)

	err := etcdclient.EtcdStore.Del(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "DeleteHPA: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", err.Error())
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "DeleteHPA: success",
	})

}

// GET 获取全局的HPA
// "/apis/v1/hpa"
func GetGlobalHPAs(c *gin.Context) {
	k8log.InfoLog("APIServer", "GetGlobalHPAs")

	key := fmt.Sprintf(serverconfig.EtcdHpaPath)

	res, err := etcdclient.EtcdStore.PrefixGet(key)

	if err != nil {
		k8log.ErrorLog("APIServer", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "GetGlobalHPAs: " + err.Error(),
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

// PUT 更新HPAStatus
func UpdateHPAStatus(c *gin.Context) {
	// log
	k8log.InfoLog("APIServer", "UpdateHPAStatus")
	hpaNamespace := c.Param(config.URL_PARAM_NAMESPACE)
	hpaName := c.Param(config.URL_PARAM_NAME)

	// 检查参数
	if hpaNamespace == "" || hpaName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}

	// 从etcd中获取指定的hpa
	key := path.Join(serverconfig.EtcdHpaPath, hpaNamespace, hpaName)
	res, err := etcdclient.EtcdStore.Get(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "UpdateHPAStatus: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", err.Error())
		return
	}
	if len(res) != 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "UpdateHPAStatus: " + "not found",
		})
		return
	}

	// 把hpaStoreJson转化为hpaStore
	hpaStore := &apiObject.HPAStore{}
	err = json.Unmarshal([]byte(res[0].Value), hpaStore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "UpdateHPAStatus: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", err.Error())
		return
	}

	// 解析请求体里面的HpaStatus
	hpaStatus := &apiObject.HPAStatus{}
	err = c.ShouldBind(hpaStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "bind hpa status err: " + err.Error(),
		})
		return
	}

	// 更新hpaStore的status
	hpaStore.Status = *hpaStatus

	// 把hpaStore转化为json
	hpaJson, err := json.Marshal(hpaStore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "marshal hpa err, " + err.Error(),
		})
		return 
	}

	// 更新etcd中的hpa
	err = etcdclient.EtcdStore.Put(key, hpaJson)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "UpdateHPAStatus: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "UpdateHPAStatus: success",
	})

}