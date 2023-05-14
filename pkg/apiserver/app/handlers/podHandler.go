package handlers

import (
	"encoding/json"
	"fmt"
	"miniK8s/pkg/apiObject"
	etcdclient "miniK8s/pkg/apiserver/app/etcdclient"
	msgutil "miniK8s/pkg/apiserver/msgUtil"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	"net/http"
	"time"

	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/util/stringutil"
	"miniK8s/util/uuid"

	"github.com/gin-gonic/gin"
)

// 获取单个Pod的信息
// "/api/v1/namespaces/:namespace/pods"
func GetPod(c *gin.Context) {
	// "/api/v1/namespaces/:namespace/pods/:name"
	// 解析里面的参数
	// namespace := c.Param("namespace")
	// name := c.Param("name")
	namespace := c.Param(config.URL_PARAM_NAMESPACE)
	name := c.Param(config.URL_PARAM_NAME)
	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}
	// 从etcd中获取
	// ETCD里面的路径是 /registry/pods/<namespace>/<pod-name>
	logStr := fmt.Sprintf("GetPod: namespace = %s, name = %s", namespace, name)
	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf("/registry/pods/%s/%s", namespace, name)
	res, err := etcdclient.EtcdStore.Get(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get pod failed " + err.Error(),
		})
		return
	}

	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "get pod err, not find pod",
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
	targetPod := res[0].Value
	c.JSON(200, gin.H{
		"data": targetPod,
	})
}

// 获取所有的Pod的信息
// API "/api/v1/namespaces/:namespace/pods"
func GetPods(c *gin.Context) {
	// namespaces := c.Param("namespace")

	namespace := c.Param(config.URL_PARAM_NAMESPACE)
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace is empty",
		})
		return
	}
	// 从etcd中获取
	// ETCD里面的路径是 /registry/pods/<namespace>/
	logStr := fmt.Sprintf("GetPods: namespace = %s", namespace)
	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf(serverconfig.EtcdPodPath+"%s/", namespace)
	res, err := etcdclient.EtcdStore.PrefixGet(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get pods failed " + err.Error(),
		})
		return
	}

	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "get pods err, not find pods",
		})
		return
	}

	// 遍历res，返回对应的Node信息
	targetPods := make([]string, 0)
	for _, pod := range res {
		targetPods = append(targetPods, pod.Value)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stringutil.StringSliceToJsonArray(targetPods),
	})
}

// POST "/api/v1/namespaces/:namespace/pods"
func AddPod(c *gin.Context) {
	// log
	k8log.InfoLog("APIServer", "AddPod")

	// 从body中获取pod的信息
	var pod apiObject.Pod
	if err := c.ShouldBind(&pod); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "parser pod failed " + err.Error(),
		})

		k8log.ErrorLog("APIServer", "AddPod: parser pod failed "+err.Error())
		return
	}

	// 检查名字是否为空
	newPodName := pod.Metadata.Name
	if newPodName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "pod name is empty",
		})
		k8log.ErrorLog("APIServer", "AddPod: pod name is empty")
		return
	}

	// 判断PodNamespace是否为空
	if pod.GetPodNamespace() == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "pod namespace is empty",
		})
		k8log.ErrorLog("APIServer", "AddPod: pod namespace is empty")
		return
	}

	// 检查name是否重复
	key := fmt.Sprintf(serverconfig.EtcdPodPath+"%s/%s", pod.GetPodNamespace(), newPodName)
	res, err := etcdclient.EtcdStore.Get(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get pod failed " + err.Error(),
		})
		return
	}

	if len(res) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "pod name has exist",
		})
		k8log.ErrorLog("APIServer", "AddPod: pod name has exist")
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "parser pod to json failed " + err.Error(),
		})
		return
	}

	// 将pod存储到etcd中
	// 持久化
	// key = stringutil.Replace(serverconfig.DefaultPod, config.URI_PARAM_NAME_PART, newPodName)

	key = fmt.Sprintf(serverconfig.EtcdPodPath+"%s/%s", pod.GetPodNamespace(), newPodName)

	// 将pod存储到etcd中
	err = etcdclient.EtcdStore.Put(key, podStoreJson)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
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
	msgutil.PublishRequestNodeScheduleMsg(podStore)
}

// "/api/v1/namespaces/:namespace/pods/:name"
func DeletePod(c *gin.Context) {
	// log
	k8log.InfoLog("APIServer", "DeletePod")

	// 解析参数
	// namespace := c.Param("namespace")
	// name := c.Param("name")
	namespace := c.Param(config.URL_PARAM_NAMESPACE)
	name := c.Param(config.URL_PARAM_NAME)

	// 检查参数是否为空
	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}

	logStr := fmt.Sprintf("DeletePod: namespace = %s, name = %s", namespace, name)
	k8log.InfoLog("APIServer", logStr)

	err := etcdclient.EtcdStore.Del(fmt.Sprintf("/registry/pods/%s/%s", namespace, name))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "delete pod failed " + err.Error(),
		})
		return
	}

	c.JSON(204, gin.H{
		"message": "delete pod success",
	})

}

// PUT "/api/v1/namespaces/:namespace/pods/:name"
func UpdatePod(c *gin.Context) {
	// TODO: 更新pod
	podName := c.Param(config.URL_PARAM_NAME)
	podNamespace := c.Param(config.URL_PARAM_NAMESPACE)

	if podName == "" || podNamespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}

	// 从etcd中获取Pod
	// 从etcd中获取
	// ETCD里面的路径是 /registry/pods/<namespace>/<pod-name>
	logStr := fmt.Sprintf("GetPod: namespace = %s, name = %s", podNamespace, podName)
	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf(serverconfig.EtcdPodPath+"%s/%s", podNamespace, podName)
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

	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get pod failed, pod is not unique",
		})
		return
	}

	// 将json转化为PodStore
	podStore := &apiObject.PodStore{}
	err = json.Unmarshal([]byte(res[0].Value), podStore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "parser json to pod failed " + err.Error(),
		})
		return
	}
	// 解析请求体里面的PodStore
	podStoreFromReq := &apiObject.PodStore{}
	err = c.ShouldBind(podStoreFromReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "parser request body to pod failed " + err.Error(),
		})
		return
	}

	// 选择性的更新Pod
	selectiveUpdatePod(podStore, podStoreFromReq)

	// 将PodStore转化为json
	podStoreJson, err := json.Marshal(podStore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "parser pod to json failed " + err.Error(),
		})
		return
	}

	// 检测GetPodNamespace和GetPodName是否为空
	if podStore.GetPodNamespace() == "" || podStore.GetPodName() == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "pod namespace or pod name is empty",
		})
		return
	}

	// 将pod存储到etcd中
	// 持久化
	key = fmt.Sprintf(serverconfig.EtcdPodPath+"%s/%s", podStore.GetPodNamespace(), podStore.GetPodName())
	err = etcdclient.EtcdStore.Put(key, podStoreJson)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "put pod to etcd failed " + err.Error(),
		})
		return
	}

	// 返回
	c.JSON(http.StatusOK, gin.H{
		"message": "update pod success",
	})
}

// "/api/v1/namespaces/:namespace/pods/:name/status"
func GetPodStatus(c *gin.Context) {
	podName := c.Param(config.URL_PARAM_NAME)
	podNamespace := c.Param(config.URL_PARAM_NAMESPACE)

	if podName == "" || podNamespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}

	// 从etcd中获取Pod
	// 从etcd中获取
	// ETCD里面的路径是 /registry/pods/<namespace>/<pod-name>
	logStr := fmt.Sprintf("GetPod: namespace = %s, name = %s", podNamespace, podName)
	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf(serverconfig.EtcdPodPath+"%s/%s", podNamespace, podName)
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

	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get pod failed, pod is not unique",
		})
		return
	}

	// 获取res[0]的值，然后转化为PodStore
	podStore := &apiObject.PodStore{}
	err = json.Unmarshal([]byte(res[0].Value), podStore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "parser pod failed " + err.Error(),
		})
		return
	}

	// 返回podStore的状态
	c.JSON(http.StatusOK, gin.H{
		"data": podStore.Status,
	})

}

// 为了区分更新Pod和更新Pod的状态，所以这里的路由是不一样的
// 理由：假如一个Pod被删除了，但是删除需要时间，如果在节点上面，Pod还在运行，就会继续回传状态信息
// 这样就会导致，删除Pod的时候，Pod又被创建了，死循环，寄中寄
// "/api/v1/namespaces/:namespace/pods/:name/status"
func UpdatePodStatus(c *gin.Context) {
	podName := c.Param(config.URL_PARAM_NAME)
	podNamespace := c.Param(config.URL_PARAM_NAMESPACE)

	if podName == "" || podNamespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}

	// 从etcd中获取Pod
	// 从etcd中获取
	// ETCD里面的路径是 /registry/pods/<namespace>/<pod-name>
	logStr := fmt.Sprintf("GetPod: namespace = %s, name = %s", podNamespace, podName)
	k8log.InfoLog("APIServer", logStr)

	key := fmt.Sprintf(serverconfig.EtcdPodPath+"%s/%s", podNamespace, podName)
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

	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get pod failed, pod is not unique",
		})
		return
	}

	// 获取res[0]的值，然后转化为PodStore
	podStore := &apiObject.PodStore{}
	err = json.Unmarshal([]byte(res[0].Value), podStore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "parser pod failed " + err.Error(),
		})
		return
	}

	// 获取请求体，转化为PodStatus
	podStatus := &apiObject.PodStatus{}
	err = c.ShouldBind(podStatus)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "parser pod status failed " + err.Error(),
		})
		return
	}

	// 更新Pod的状态
	selectiveUpdatePodStatus(podStore, podStatus)

	// 将PodStore转化为json
	podStoreJson, err := json.Marshal(podStore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "parser pod failed " + err.Error(),
		})
		return
	}

	// 更新Pod
	err = etcdclient.EtcdStore.Put(key, podStoreJson)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "update pod failed " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "update pod success",
	})
}

// 选择性的更新Pod的状态
func selectiveUpdatePodStatus(oldPod *apiObject.PodStore, podStatus *apiObject.PodStatus) {
	// 根据podStatus的值，更新apiObject的值
	if podStatus.Phase != "" {
		oldPod.Status.Phase = podStatus.Phase
	}

	if podStatus.PodIP != "" {
		oldPod.Status.PodIP = podStatus.PodIP
	}

	// UpdateTime
	oldPod.Status.UpdateTime = time.Now()

}

func selectiveUpdatePod(oldPod *apiObject.PodStore, putPod *apiObject.PodStore) {
	// Labels处理
	if len(putPod.Metadata.Labels) != 0 {
		// 先清空原来的Labels
		oldPod.Metadata.Labels = make(map[string]string)
		for key, value := range putPod.Metadata.Labels {
			oldPod.Metadata.Labels[key] = value
		}
	}

	// Annotations处理
	if len(putPod.Metadata.Annotations) != 0 {
		// 先清空原来的Annotations
		oldPod.Metadata.Annotations = make(map[string]string)
		for key, value := range putPod.Metadata.Annotations {
			oldPod.Metadata.Annotations[key] = value
		}
	}

	// Spec暂时不可以更新
	// Status选择性更新
	selectiveUpdatePodStatus(oldPod, &(putPod.Status))

}
