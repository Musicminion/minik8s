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
	"path"

	"github.com/gin-gonic/gin"
)

// GetWorkFlow
// GET /apis/v1/namespaces/:namespace/workflows/:workflow
func GetWorkFlow(c *gin.Context) {
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

	logStr := "GetWorkFlow: namespace=" + namespace + ", name=" + name
	k8log.InfoLog("APIServer", logStr)

	key := path.Join(serverconfig.EtcdWorkflowPath, namespace, name)
	res, err := etcdclient.EtcdStore.Get(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "GetWorkFlow: " + err.Error(),
		})
		return
	}

	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "GetWorkFlow: not found",
		})
		return
	}

	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "GetWorkFlow: more than one result",
		})
		return
	}

	targetWorkFlow := res[0].Value

	c.JSON(http.StatusOK, gin.H{
		"data": targetWorkFlow,
	})
}

// GetWorkFlows
// GET /apis/v1/namespaces/:namespace/workflows
func GetWorkFlows(c *gin.Context) {
	namespace := c.Param(config.URL_PARAM_NAMESPACE)
	if namespace == "" {
		namespace = config.DefaultNamespace
	}

	logStr := "GetWorkFlows: namespace=" + namespace
	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf(serverconfig.EtcdWorkflowPath+"%s/", namespace)

	res, err := etcdclient.EtcdStore.PrefixGet(key)

	if err != nil {
		k8log.ErrorLog("APIServer", "GetWorkFlows: "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "GetWorkFlows: " + err.Error(),
		})
		return
	}

	targetWorkFlows := make([]string, 0)

	for _, v := range res {
		targetWorkFlows = append(targetWorkFlows, v.Value)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stringutil.StringSliceToJsonArray(targetWorkFlows),
	})
}

// AddWorkFlow
// POST /apis/v1/namespaces/:namespace/workflows
func AddWorkFlow(c *gin.Context) {
	// log
	k8log.InfoLog("APIServer", "AddWorkFlow")

	// 从请求中获取workflow
	var workflow apiObject.Workflow

	if err := c.ShouldBindJSON(&workflow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "AddWorkFlow: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "AddWorkFlow: "+err.Error())
		return
	}

	newWorkflowName := workflow.Metadata.Name

	if newWorkflowName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "AddWorkFlow: name is empty",
		})
		k8log.ErrorLog("APIServer", "AddWorkFlow: name is empty")
		return
	}

	// 检查namespace
	if workflow.Metadata.Namespace == "" {
		workflow.Metadata.Namespace = config.DefaultNamespace
	}

	// 检查是否存在
	key := fmt.Sprintf(serverconfig.EtcdWorkflowPath+"%s/%s", workflow.Metadata.Namespace, newWorkflowName)

	res, err := etcdclient.EtcdStore.Get(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "AddWorkFlow: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "AddWorkFlow: "+err.Error())
		return
	}

	if len(res) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "AddWorkFlow: already exists",
		})
		k8log.ErrorLog("APIServer", "AddWorkFlow: already exists")
		return
	}

	// 添加
	workflow.Metadata.UUID = uuid.NewUUID()

	// 把workflow转换成workflowStore
	workFlowStore := workflow.ToWorkflowStore()

	workFlowStoreJson, err := json.Marshal(workFlowStore)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "AddWorkFlow: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "AddWorkFlow: "+err.Error())
		return
	}

	key = fmt.Sprintf(serverconfig.EtcdWorkflowPath+"%s/%s", workflow.Metadata.Namespace, newWorkflowName)

	err = etcdclient.EtcdStore.Put(key, workFlowStoreJson)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "AddWorkFlow: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "AddWorkFlow: "+err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "AddWorkFlow: success",
	})

	/*
	   后面如果要做什么再加
	*/
}

// DeleteWorkFlow
// DELETE /apis/v1/namespaces/:namespace/workflows/:workflow
func DeleteWorkFlow(c *gin.Context) {
	// log
	k8log.InfoLog("APIServer", "DeleteWorkFlow")

	namespace := c.Param(config.URL_PARAM_NAMESPACE)
	name := c.Param(config.URL_PARAM_NAME)

	if namespace == "" {
		namespace = config.DefaultNamespace
	}

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "DeleteWorkFlow: name is empty",
		})
		k8log.ErrorLog("APIServer", "DeleteWorkFlow: name is empty")
		return
	}

	key := fmt.Sprintf(serverconfig.EtcdWorkflowPath+"%s/%s", namespace, name)

	k8log.InfoLog("APIServer", "DeleteWorkFlow: key="+key)

	err := etcdclient.EtcdStore.Del(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "DeleteWorkFlow: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "DeleteWorkFlow: "+err.Error())
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "DeleteWorkFlow: success",
	})
}

// UpdateWorkFlow
// PUT /apis/v1/namespaces/:namespace/workflows/:workflow
func UpdateWorkFlow(c *gin.Context) {
	// Not Supported
}

// GetWorkFlowStatus
// GET /apis/v1/namespaces/:namespace/workflows/:workflow/status
func GetWorkFlowStatus(c *gin.Context) {
	name := c.Param(config.URL_PARAM_NAME)
	namespace := c.Param(config.URL_PARAM_NAMESPACE)

	if name == "" || namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "GetWorkFlowStatus: name or namespace is empty",
		})
		return
	}

	logStr := fmt.Sprintf("GetWorkFlowStatus: name=%s, namespace=%s", name, namespace)

	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf(serverconfig.EtcdWorkflowPath+"%s/%s", namespace, name)

	res, err := etcdclient.EtcdStore.Get(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "GetWorkFlowStatus: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "GetWorkFlowStatus: "+err.Error())
		return
	}

	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "GetWorkFlowStatus: not found",
		})
		k8log.ErrorLog("APIServer", "GetWorkFlowStatus: not found")
		return
	}

	// 解析
	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "GetWorkFlowStatus: res len != 1",
		})
		k8log.ErrorLog("APIServer", "GetWorkFlowStatus: res len != 1")
		return
	}

	workflowStore := &apiObject.WorkflowStore{}
	err = json.Unmarshal([]byte(res[0].Value), workflowStore)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "GetWorkFlowStatus: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "GetWorkFlowStatus: "+err.Error())
		return
	}

	// 转换成workflow
	c.JSON(http.StatusOK, gin.H{
		"data": workflowStore.Status,
	})

}

// UpdateWorkFlowStatus
// PUT /apis/v1/namespaces/:namespace/workflows/:workflow/status
func UpdateWorkFlowStatus(c *gin.Context) {

	name := c.Param(config.URL_PARAM_NAME)
	namespace := c.Param(config.URL_PARAM_NAMESPACE)

	if name == "" || namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "UpdateWorkFlowStatus: name or namespace is empty",
		})
		return
	}

	logStr := fmt.Sprintf("UpdateWorkFlowStatus: namespace=%s, name=%s", namespace, name)
	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf(serverconfig.EtcdWorkflowPath+"%s/%s", namespace, name)

	res, err := etcdclient.EtcdStore.Get(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "UpdateWorkFlowStatus: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "UpdateWorkFlowStatus: "+err.Error())
		return
	}

	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "UpdateWorkFlowStatus: not found",
		})
		k8log.ErrorLog("APIServer", "UpdateWorkFlowStatus: not found")
		return
	}

	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "UpdateWorkFlowStatus: res len != 1",
		})
		k8log.ErrorLog("APIServer", "UpdateWorkFlowStatus: res len != 1")
		return
	}

	workflowStore := &apiObject.WorkflowStore{}
	err = json.Unmarshal([]byte(res[0].Value), workflowStore)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "UpdateWorkFlowStatus: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "UpdateWorkFlowStatus: "+err.Error())
		return
	}

	flowStatus := &apiObject.WorkflowStatus{}
	err = c.ShouldBind(flowStatus)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "UpdateWorkFlowStatus: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "UpdateWorkFlowStatus: "+err.Error())
		return
	}

	selectiveUpdateFlowStatus(workflowStore, flowStatus)

	workFlowStoreJson, err := json.Marshal(workflowStore)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "UpdateWorkFlowStatus: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "UpdateWorkFlowStatus: "+err.Error())
		return
	}

	err = etcdclient.EtcdStore.Put(key, workFlowStoreJson)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "UpdateWorkFlowStatus: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "UpdateWorkFlowStatus: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "UpdateWorkFlowStatus: success",
	})

}

// GET 获取全局的workflow
// GET /apis/v1/workflows
func GetGlobalWorkFlows(c *gin.Context) {
	logStr := "GetGlobalWorkFlows"
	k8log.InfoLog("APIServer", logStr)

	key := serverconfig.EtcdWorkflowPath
	res, err := etcdclient.EtcdStore.PrefixGet(key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "GetGlobalWorkFlows: " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "GetGlobalWorkFlows: "+err.Error())
		return
	}
	targetFlows := make([]string, 0)

	for _, v := range res {
		targetFlows = append(targetFlows, v.Value)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stringutil.StringSliceToJsonArray(targetFlows),
	})
}

func selectiveUpdateFlowStatus(oldStatus *apiObject.WorkflowStore, newStatus *apiObject.WorkflowStatus) {
	if newStatus.Phase != "" {
		oldStatus.Status.Phase = newStatus.Phase
	}
	if newStatus.Result != "" {
		oldStatus.Status.Result = newStatus.Result
	}
}
