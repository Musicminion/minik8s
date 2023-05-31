package handlers

import (
	"encoding/json"
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/pkg/message"
	"net/http"
	"path"
	"strconv"

	"miniK8s/pkg/apiserver/app/etcdclient"
	"miniK8s/pkg/apiserver/app/helper"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/k8log"
	"miniK8s/util/stringutil"
	"miniK8s/util/uuid"

	"github.com/gin-gonic/gin"
)

// 添加新的Service
// POST "/api/v1/namespaces/:namespace/services"
func AddService(c *gin.Context) {
	k8log.InfoLog("APIServer", "AddService: add new service")
	// POST请求，获取请求体
	var service apiObject.Service
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "parser service failed " + err.Error(),
		})

		k8log.ErrorLog("APIServer", "AddService: parser service failed "+err.Error())
		return
	}

	// 检查name是否重复
	key := fmt.Sprintf(serverconfig.EtcdServicePath+"%s/%s", service.Metadata.Namespace, service.Metadata.Name)
	res, err := etcdclient.EtcdStore.PrefixGet(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get service failed " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "AddService: get service failed "+err.Error())
		return
	}

	if len(res) != 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "service name already exist",
		})
		k8log.ErrorLog("APIServer", "AddService: service name already exist")
		return
	}
	// 检查Service的kind是否正确
	if service.Kind != "Service" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "service kind is not Service",
		})
		k8log.ErrorLog("APIServer", "AddService: service kind is not Service")
		return
	}

	// 为service分配IP
	if service.Spec.ClusterIP != "" {
		// 已有IP，检验合法性
		err = helper.JudgeServiceIPAddress(service.Spec.ClusterIP)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "service ip address is not valid",
			})
			return
		}
	} else {
		service.Spec.ClusterIP, err = helper.AllocClusterIP()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "alloc cluster ip failed",
			})
			return
		}
	}

	// 给Service设置UUID, 所以哪怕用户故意设置UUID也会被覆盖
	service.Metadata.UUID = uuid.NewUUID()

	// 将Service转化为ServiceStore
	serviceStore := service.ToServiceStore()

	// 把serviceStore转化为json
	serviceJson, err := json.Marshal(serviceStore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "service marshal to json failed" + err.Error(),
		})
		return
	}

	serviceUpdate := &entity.ServiceUpdate{
		Action:        message.CREATE,
		ServiceTarget: *serviceStore,
	}

	for key, value := range service.Spec.Selector {
		func(key, value string) {
			// 替换可变参 namespace
			svcSelectorURL := path.Join(serverconfig.EtcdServiceSelectorPath, key, value, service.Metadata.UUID)

			// 为每个service的每个selector创建一个etcd url，方便后续的查找
			if err := etcdclient.EtcdStore.Put(svcSelectorURL, serviceJson); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "add service to etcd failed" + err.Error(),
				})
				return
			}
			var endpoints []apiObject.Endpoint
			if endpoints, err = helper.GetEndpoints(key, value); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "get endpoints failed" + err.Error(),
				})
				return
			} else {
				// 添加Endpoints到service
				serviceStore.Status.Endpoints = append(serviceStore.Status.Endpoints, endpoints...)
			}

			k8log.DebugLog("APIServer", "endpoints number of service "+service.GetObjectName()+" is "+strconv.Itoa(len(serviceUpdate.ServiceTarget.Status.Endpoints)))

		}(key, value)
	}
	// 更新serviceUpdate，并更新etcd中的serviceStore
	serviceUpdate.ServiceTarget = *serviceStore
	serviceJson, err = json.Marshal(serviceStore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "service marshal to json failed" + err.Error(),
		})
		return
	}

	// 将Service信息写入etcd
	key = fmt.Sprintf(serverconfig.EtcdServicePath+"%s/%s", service.Metadata.Namespace, service.Metadata.Name)
	err = etcdclient.EtcdStore.Put(key, serviceJson)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "put service to etcd failed" + err.Error(),
		})
		return
	}
	// 返回201处理成功
	c.JSON(http.StatusCreated, gin.H{
		"message": "create service success",
	})

	// publishServiceUpdate(serviceUpdate)
	k8log.DebugLog("APIServer", "AddService: serviceUpdate")
	message.PublishUpdateService(serviceUpdate)

}

// 获取单个Service信息
// 某个特定的Service状态 对应的ServiceSpecURL = "/api/v1/services/:name"
func GetService(c *gin.Context) {
	// 尝试解析请求里面的name
	name := c.Param("name")
	namespace := c.Param("namespace")
	logStr := "GetSerive: name = " + name
	k8log.InfoLog("APIServer", logStr)

	// 如果解析成功，返回对应的Service信息
	if name != "" {
		res, err := etcdclient.EtcdStore.PrefixGet(path.Join(serverconfig.EtcdServicePath, namespace, name))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "get service failed " + err.Error(),
			})
			return
		}
		// 没找到
		if len(res) == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "get service err, not find service",
			})
			return
		}

		// 处理res，如果发现有多个Service，返回错误
		if len(res) != 1 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "get service err, find more than one service",
			})
			return
		}
		// 遍历res，返回对应的Service信息
		targetService := res[0].Value
		c.JSON(http.StatusOK, gin.H{
			"data": targetService,
		})
		return
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "name is empty",
		})
		return
	}
}

// 获取所有Service信息
func GetServices(c *gin.Context) {
	// TODO: 根据namespace查出所有的Service
	namespace := c.Param(config.URL_PARAM_NAMESPACE)
	if namespace == "" {
		namespace = config.DefaultNamespace
	}

	res, err := etcdclient.EtcdStore.PrefixGet(serverconfig.EtcdServicePath + namespace + "/")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "get services failed " + err.Error(),
		})
		return
	}
	// 遍历res，返回对应的Service信息
	var services []string
	for _, service := range res {
		services = append(services, service.Value)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stringutil.StringSliceToJsonArray(services),
	})
}

// 删除Service信息
func DeleteService(c *gin.Context) {
	// 尝试解析请求里面的name和namespace
	name := c.Params.ByName("name")
	namespace := c.Params.ByName("namespace")
	service := apiObject.ServiceStore{}
	// 如果解析成功，删除对应的Service信息
	if name != "" {
		// log
		logStr := "DeleteService: name = " + name
		k8log.InfoLog("APIServer", logStr)

		// 从etcd中获取
		// ETCD里面的路径是 /registry/services/<namespace>/<pod-name>
		key := fmt.Sprintf(serverconfig.EtcdServicePath+"%s/%s", namespace, name)
		k8log.DebugLog("APIServer", "DeleteService: path: "+key)
		res, err := etcdclient.EtcdStore.Get(key)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "get service failed " + err.Error(),
			})
			return
		}
		if len(res) == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "service not found",
			})
			return
		}
		if len(res) != 1 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "get service failed, service is not unique",
			})
			return
		}
		// 将json转化为PodStore
		err = json.Unmarshal([]byte(res[0].Value), &service)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "parser json to service failed " + err.Error(),
			})
			return
		}
		// 删除service
		err = etcdclient.EtcdStore.Del(key)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "delete service failed " + err.Error(),
			})
			return
		}
		// 删除service的所有Label
		for key, value := range service.Spec.Selector {
			k8log.DebugLog("APIServer", "DeleteService: delete service label: "+key+" "+value)
			err = etcdclient.EtcdStore.PrefixDel(path.Join(serverconfig.EtcdServiceSelectorPath, key, value, service.Metadata.UUID))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "delete service failed " + err.Error(),
				})
				return
			}
		}

	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "name is empty",
		})
		return
	}

	serviceUpdate := &entity.ServiceUpdate{
		Action:        message.DELETE,
		ServiceTarget: service,
	}

	message.PublishUpdateService(serviceUpdate)

	c.JSON(http.StatusNoContent, gin.H{
		"message": "delete service success",
	})
}

// 更新Service信息
// PUT "/api/v1/namespaces/:namespace/services/:name"
func UpdateService(c *gin.Context) {
	serviceName := c.Param(config.URL_PARAM_NAME)
	serviceNamespace := c.Param(config.URL_PARAM_NAMESPACE)

	if serviceNamespace == "" {
		serviceNamespace = config.DefaultNamespace
	}

	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "name is empty",
		})
		return
	}

	// 从etcd中获取Service
	// 从etcd中获取
	// ETCD里面的路径是 /registry/services/<namespace>/<pod-name>
	logStr := fmt.Sprintf("GetPod: namespace = %s, name = %s", serviceNamespace, serviceName)
	k8log.InfoLog("APIServer", logStr)

	key := path.Join(serverconfig.EtcdServicePath, serviceNamespace, serviceName)
	res, err := etcdclient.EtcdStore.Get(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get service failed " + err.Error(),
		})
		return
	}

	if len(res) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "service not found",
		})
		return
	}

	if len(res) != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "get service failed, service is not unique",
		})
		return
	}

	// 将json转化为PodStore
	serviceStore := &apiObject.ServiceStore{}
	err = json.Unmarshal([]byte(res[0].Value), serviceStore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "parser json to service failed " + err.Error(),
		})
		return
	}

	// 解析请求体里面的ServiceStore
	serviceStoreFromReq := &apiObject.ServiceStore{}
	err = c.ShouldBind(serviceStoreFromReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "parser request body to service failed " + err.Error(),
		})
		return
	}

	// 选择性的更新Service
	selectiveUpdatePService(serviceStore, serviceStoreFromReq)

	// 将ServiceStore转化为json
	serviceStoreJson, err := json.Marshal(serviceStore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "parser service to json failed " + err.Error(),
		})
		return
	}

	// 检测GetPodNamespace和GetPodName是否为空
	if serviceStore.GetNamespace() == "" || serviceStore.GetName() == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "service namespace or service name is empty",
		})
		return
	}

	// 将service存储到etcd中
	// 持久化
	key = fmt.Sprintf(serverconfig.EtcdPodPath+"%s/%s", serviceStore.GetNamespace(), serviceStore.GetName())
	err = etcdclient.EtcdStore.Put(key, serviceStoreJson)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "put service to etcd failed " + err.Error(),
		})
		return
	}

	// 返回
	c.JSON(http.StatusOK, gin.H{
		"message": "update service success",
	})
}

// 选择性的更新Pod
func selectiveUpdatePService(oldSerivce *apiObject.ServiceStore, newSerivce *apiObject.ServiceStore) {
	// Labels处理
	if len(newSerivce.Metadata.Labels) != 0 {
		// 先清空原来的Labels与Endpoints
		// 遍历Labels
		for key, value := range newSerivce.Metadata.Labels {
			// 删除所有匹配的Endpoints
			etcdclient.EtcdStore.PrefixDel(path.Join(serverconfig.EndpointPath, key, value))
		}
		oldSerivce.Metadata.Labels = make(map[string]string)
		for key, value := range newSerivce.Metadata.Labels {
			oldSerivce.Metadata.Labels[key] = value
		}
	}

	// Annotations处理
	if len(newSerivce.Metadata.Annotations) != 0 {
		// 先清空原来的Annotations
		newSerivce.Metadata.Annotations = make(map[string]string)
		for key, value := range newSerivce.Metadata.Annotations {
			newSerivce.Metadata.Annotations[key] = value
		}
	}

	// Spec暂时不可以更新
	// Status选择性更新
	selectiveUpdateServiceStatus(oldSerivce, &(newSerivce.Status))
}

// 选择性的更新ServiceStatus
func selectiveUpdateServiceStatus(oldSerivce *apiObject.ServiceStore, newStatus *apiObject.ServiceStatus) {

}
