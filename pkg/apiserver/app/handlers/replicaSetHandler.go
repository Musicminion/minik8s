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

// "/apis/v1/namespaces/:namespace/replicasets/:name"
// GET
func GetReplicaSet(c *gin.Context) {
	// 获取指定的replicaSet

	// 解析里面的参数
	namespace := c.Param(config.URL_PARAM_NAMESPACE)
	name := c.Param(config.URL_PARAM_NAME)

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}

	logStr := "GetReplicaSet: namespace=" + namespace + ", name=" + name

	k8log.InfoLog("APIServer", logStr)

	// 从etcd中获取指定的replicaSet
	// 完整路径 /registry/replicasets/:namespace/:name

	key := fmt.Sprintf(serverconfig.EtcdReplicaSetPath+"%s/%s", namespace, name)

	// 从etcd中获取指定的replicaSet
	res, err := etcdclient.EtcdStore.Get(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "replicaSet not found",
		})
		return
	}

	if len(res) > 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "more than one replicaSet found",
		})
		return
	}

	targetReplicaset := res[0].Value

	c.JSON(http.StatusOK, gin.H{
		"data": targetReplicaset,
	})
}

// "/apis/v1/namespaces/:namespace/replicasets"
// GET
func GetReplicaSets(c *gin.Context) {
	// 获取指定的所有的replicaSet

	// 解析里面的参数
	namespace := c.Param(config.URL_PARAM_NAMESPACE)

	if namespace == "" {
		namespace = config.DefaultNamespace
	}

	logStr := "GetReplicaSets: namespace=" + namespace

	k8log.InfoLog("APIServer", logStr)

	// 从etcd中获取指定的replicaSet
	key := fmt.Sprintf(serverconfig.EtcdReplicaSetPath+"%s", namespace)
	res, err := etcdclient.EtcdStore.PrefixGet(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "replicaSets not found",
		})
		return
	}

	targetReplicaseta := make([]string, 0)

	for _, v := range res {
		targetReplicaseta = append(targetReplicaseta, v.Value)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stringutil.StringSliceToJsonArray(targetReplicaseta),
	})

}

// "/apis/v1/replicasets"
// GET
func GetGlobalReplicaSets(c *gin.Context) {
	key := fmt.Sprintf(serverconfig.EtcdReplicaSetPath)
	res, err := etcdclient.EtcdStore.PrefixGet(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	targetReplicaseta := make([]string, 0)

	for _, v := range res {
		targetReplicaseta = append(targetReplicaseta, v.Value)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stringutil.StringSliceToJsonArray(targetReplicaseta),
	})
}

// "/apis/v1/namespaces/:namespace/replicasets"
// POST
func AddReplicaSet(c *gin.Context) {
	// log
	k8log.InfoLog("APIServer", "AddReplicaSet")

	// 解析里面的参数
	var replicaSet apiObject.ReplicaSet

	if err := c.ShouldBindJSON(&replicaSet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		k8log.ErrorLog("APIServer", err.Error())
		return
	}

	// 检查ReplicaSet的合法性
	newReplicasetName := replicaSet.Metadata.Name

	if newReplicasetName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "replicaSet name is empty",
		})
		return
	}

	// 检查replicaSet是否存在
	key := fmt.Sprintf(serverconfig.EtcdReplicaSetPath+"%s/%s", replicaSet.GetReplicaSetNamespace(), replicaSet.GetReplicaSetName())

	// 从etcd中获取指定的replicaSet
	res, err := etcdclient.EtcdStore.Get(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if len(res) != 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error": "replicaSet already exists",
		})
		k8log.ErrorLog("APIServer", "replicaSet already exists")
		return
	}

	// 给replicaSet添加UUID, 用于后面的更新
	// 哪怕用户自己指定了UUID, 也会被覆盖
	replicaSet.Metadata.UUID = uuid.NewUUID()

	// 创建replicaSet
	replicaSetStore := replicaSet.ToReplicaSetStore()

	replicaSetStoreJson, err := json.Marshal(replicaSetStore)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 找到key
	key = fmt.Sprintf(serverconfig.EtcdReplicaSetPath+"%s/%s", replicaSet.GetReplicaSetNamespace(), newReplicasetName)

	// 创建replicaSet
	err = etcdclient.EtcdStore.Put(key, replicaSetStoreJson)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "replicaSet created",
	})

	k8log.InfoLog("APIServer", "replicaSet created")

	/*
		后面的处理逻辑待定
	*/
}

// "/apis/v1/namespaces/:namespace/replicasets/:name"
// DELETE
func DeleteReplicaSet(c *gin.Context) {
	// 删除指定的replicaSet

	// 解析里面的参数
	namespace := c.Param(config.URL_PARAM_NAMESPACE)
	name := c.Param(config.URL_PARAM_NAME)

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}

	logStr := "DeleteReplicaSet: namespace=" + namespace + ", name=" + name

	k8log.InfoLog("APIServer", logStr)

	// 组装key
	key := fmt.Sprintf(serverconfig.EtcdReplicaSetPath+"%s/%s", namespace, name)

	err := etcdclient.EtcdStore.Del(key)

	if err != nil {
		k8log.DebugLog("APIServer", "delete replicaset err, "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusNoContent, gin.H{
		"message": "replicaSet deleted",
	})

	k8log.InfoLog("APIServer", "replicaSet deleted")

}

// "/apis/v1/namespaces/:namespace/replicasets/:name/status"
// GET
func GetReplicaSetStatus(c *gin.Context) {
	// 获取指定replicaSet的状态
	replicaNamespace := c.Param(config.URL_PARAM_NAMESPACE)
	replicaName := c.Param(config.URL_PARAM_NAME)

	if replicaNamespace == "" || replicaName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}

	// 从etcd中获取指定的replicaSet
	logStr := "GetReplicaSetStatus: namespace=" + replicaNamespace + ", name=" + replicaName

	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf(serverconfig.EtcdReplicaSetPath+"%s/%s", replicaNamespace, replicaName)

	res, err := etcdclient.EtcdStore.Get(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get replicaset err, " + err.Error(),
		})
		return
	}

	if len(res) != 1 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "replicaSet not found",
		})
		return
	}

	// 解析replicaSet
	replicaStore := &apiObject.ReplicaSetStore{}

	err = json.Unmarshal([]byte(res[0].Value), replicaStore)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unmarshal replicaset err, " + err.Error(),
		})
		return
	}

	// 获取replicaSet的状态

	c.JSON(http.StatusOK, gin.H{
		"data": replicaStore.Status,
	})

}

// "/apis/v1/namespaces/:namespace/replicasets/:name/status"
// PUT
func UpdateReplicaSetStatus(c *gin.Context) {
	// 更新指定replicaSet的状态

	replicaNamespace := c.Param(config.URL_PARAM_NAMESPACE)
	replicaName := c.Param(config.URL_PARAM_NAME)

	if replicaNamespace == "" || replicaName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}

	// 从etcd中获取指定的replicaSet
	logStr := "UpdateReplicaSetStatus: namespace=" + replicaNamespace + ", name=" + replicaName
	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf(serverconfig.EtcdReplicaSetPath+"%s/%s", replicaNamespace, replicaName)

	res, err := etcdclient.EtcdStore.Get(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get replicaset err, " + err.Error(),
		})
		return
	}

	if len(res) != 1 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "replicaSet not found",
		})
		return
	}

	// 获取res[0].Value
	oldReplicaSet := &apiObject.ReplicaSetStore{}
	err = json.Unmarshal([]byte(res[0].Value), oldReplicaSet)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unmarshal replicaset err, " + err.Error(),
		})
		return
	}

	// 解析请求体里面的replicaSet
	replicaStatus := &apiObject.ReplicaSetStatus{}
	err = c.ShouldBind(replicaStatus)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "bind replicaset err, " + err.Error(),
		})
		return
	}

	// 选择性的更新replicaSet的状态
	selectiveUpdateReplicaStatus(oldReplicaSet, replicaStatus)

	// replicaSet转化为json
	replicaSetJson, err := json.Marshal(oldReplicaSet)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "marshal replicaset err, " + err.Error(),
		})
		return
	}

	err = etcdclient.EtcdStore.Put(key, replicaSetJson)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "update replicaset err, " + err.Error(),
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"message": "replicaSet status updated",
	})

}

// 更新replicaSet的状态的时候，请务必完整的传递replicaSet的状态，不然会导致状态丢失

func selectiveUpdateReplicaStatus(oldReplica *apiObject.ReplicaSetStore, newReplicaStatus *apiObject.ReplicaSetStatus) {

	oldReplica.Status.Replicas = newReplicaStatus.Replicas
	oldReplica.Status.Conditions = newReplicaStatus.Conditions
	oldReplica.Status.ReadyReplicas = newReplicaStatus.ReadyReplicas

	if len(newReplicaStatus.Conditions) > 0 {
		oldReplica.Status.Conditions = newReplicaStatus.Conditions
	}

}

// UpdateReplicaSet
// PUT
// "/apis/v1/namespaces/:namespace/replicasets/:name"
func UpdateReplicaSet(c *gin.Context) {
	k8log.InfoLog("APIServer", "UpdateReplicaSet")

	name := c.Param(config.URL_PARAM_NAME)
	namespace := c.Param(config.URL_PARAM_NAMESPACE)

	if namespace == "" {
		namespace = config.DefaultNamespace
	}

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "name is empty",
		})
		return
	}

	// 从etcd中获取指定的replicaSet
	logStr := "UpdateReplicaSet: namespace=" + namespace + ", name=" + name
	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf(serverconfig.EtcdReplicaSetPath+"%s/%s", namespace, name)
	res, err := etcdclient.EtcdStore.Get(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get replicaset err, " + err.Error(),
		})
		return
	}

	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "replicaSet not found",
		})
		return
	}

	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "replicaSet is not unique",
		})
		return
	}

	replicaStore := &apiObject.ReplicaSetStore{}
	err = json.Unmarshal([]byte(res[0].Value), replicaStore)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unmarshal replicaset err, " + err.Error(),
		})
		return
	}

	// 解析请求体里面的replicaSet
	replicaSet := &apiObject.ReplicaSetStore{}
	err = c.ShouldBind(replicaSet)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "bind replicaset err, " + err.Error(),
		})
		return
	}

	// 选择性的更新replicaSet
	selectiveUpdateReplicaSet(replicaStore, replicaSet)

	// replicaSet转化为json
	replicaSetJson, err := json.Marshal(replicaStore)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "marshal replicaset err, " + err.Error(),
		})
		return
	}

	key = fmt.Sprintf(serverconfig.EtcdReplicaSetPath+"%s/%s", replicaStore.Metadata.Namespace, replicaStore.Metadata.Name)
	err = etcdclient.EtcdStore.Put(key, replicaSetJson)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "update replicaset err, " + err.Error(),
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"message": "replicaSet updated",
	})
}

// 选择性的更新replicaSet
func selectiveUpdateReplicaSet(oldReplica *apiObject.ReplicaSetStore, newReplica *apiObject.ReplicaSetStore) {
	oldReplica.Spec.Replicas = newReplica.Spec.Replicas
}
