package handlers

import (
	"encoding/json"
	"miniK8s/pkg/apiObject"
	"strconv"

	"miniK8s/pkg/apiserver/app/etcdclient"
	msgutil "miniK8s/pkg/apiserver/msgUtil"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/config"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/k8log"
	"miniK8s/util/stringutil"
	"miniK8s/util/uuid"
	"path"

	"github.com/gin-gonic/gin"
)

// 添加新的Service
// POST "/api/v1/namespaces/:namespace/services"
func AddService(c *gin.Context) {
	// log
	k8log.InfoLog("APIServer", "AddService: add new service")
	// POST请求，获取请求体
	var service apiObject.Service
	if err := c.ShouldBind(&service); err != nil {
		c.JSON(500, gin.H{
			"error": "parser service failed " + err.Error(),
		})

		k8log.ErrorLog("APIServer", "AddService: parser service failed "+err.Error())
		return
	}

	// 检查name是否重复
	res, err := etcdclient.EtcdStore.PrefixGet(serverconfig.EtcdServicePath + service.Metadata.Name)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "get service failed " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "AddService: get service failed "+err.Error())
		return
	}

	if len(res) != 0 {
		c.JSON(500, gin.H{
			"error": "service name already exist",
		})
		k8log.ErrorLog("APIServer", "AddService: service name already exist")
		return
	}
	// 检查Service的kind是否正确
	if service.Kind != "Service" {
		c.JSON(500, gin.H{
			"error": "service kind is not Service",
		})
		k8log.ErrorLog("APIServer", "AddService: service kind is not Service")
		return
	}

	// 给Service设置UUID, 所以哪怕用户故意设置UUID也会被覆盖
	service.Metadata.UUID = uuid.NewUUID()

	// 将Service转化为ServiceStore
	serviceStore := service.ToServiceStore()

	// 把serviceStore转化为json
	serviceJson, err := json.Marshal(serviceStore)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "service marshal to json failed" + err.Error(),
		})
		return
	}

	// 将Service信息写入etcd
	etcdURL := serverconfig.EtcdServicePath + service.Metadata.Name
	err = etcdclient.EtcdStore.Put(etcdURL, serviceJson)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "put service to etcd failed" + err.Error(),
		})
		return
	}
	// 返回201处理成功
	c.JSON(201, gin.H{
		"message": "create service success",
	})

	serviceUpdate := &entity.ServiceUpdate{
		Action: entity.CREATE,
		ServiceTarget: entity.ServiceWithEndpoints{
			Service:   service,
			Endpoints: make([]apiObject.Endpoint, 0),
		},
	}

	// TODO: scan etcd and find all endpoints of this service
	for key, value := range service.Spec.Selector {
		func(key, value string) {
			// 替换可变参 namespace
			etcdURL := path.Join(config.ServiceURL, key, value, service.Metadata.UUID)
			etcdURL = stringutil.Replace(etcdURL, config.URL_PARAM_NAMESPACE_PART, service.GetNamespace())
			res, err := etcdclient.EtcdStore.PrefixGet(serverconfig.EtcdPodPath)
			// if err := etcdclient.EtcdStore.Put(etcdURL, serviceJson); err != nil {
			// 	c.JSON(500, gin.H{
			// 		"error": "add service to etcd failed" + err.Error(),
			// 	})
			// 	return
			// }
			// if endpoints, err := helper.GetEndpoints(key, value); err != nil {
			// 	c.JSON(500, gin.H{
			// 		"error": "get endpoints failed" + err.Error(),
			// 	})
			// 	return
			// } else {
			// 	serviceUpdate.ServiceTarget.Endpoints = append(serviceUpdate.ServiceTarget.Endpoints, endpoints...)
			// }

			if err != nil {
				c.JSON(500, gin.H{
					"error": "get endpoints failed" + err.Error(),
				})
				return
			} else {
				// iterate res
				for _, r := range res {
					// pod to endpoint
					pod := &apiObject.PodStore{}
					if err := json.Unmarshal([]byte(r.Value), pod); err != nil {
						c.JSON(500, gin.H{
							"error": "unmarshal pod failed" + err.Error(),
						})
						return
					}
					endpoint := apiObject.Endpoint{}
					endpoint.Metadata.Name = pod.GetPodName()
					endpoint.Metadata.Namespace = pod.GetPodNamespace()
					endpoint.Metadata.UUID = pod.Metadata.UUID
					endpoint.Metadata.Labels = pod.Metadata.Labels
					endpoint.Metadata.Annotations = pod.Metadata.Annotations
					endpoint.IP = pod.Status.PodIP
					// add endpoint to serviceUpdate.ServiceTarget.Endpoints
					serviceUpdate.ServiceTarget.Endpoints = append(serviceUpdate.ServiceTarget.Endpoints, endpoint)
				}
				k8log.DebugLog("APIServer", "endpoints number of service "+service.GetName()+" is "+strconv.Itoa(len(serviceUpdate.ServiceTarget.Endpoints)))
				// serviceUpdate.ServiceTarget.Endpoints = append(serviceUpdate.ServiceTarget.Endpoints, endpoints...)
			}
		}(key, value)
	}
	// publishServiceUpdate(serviceUpdate)
	msgutil.PublishUpdateService(serviceUpdate)

}

// 获取单个Service信息
// 某个特定的Service状态 对应的ServiceSpecURL = "/api/v1/services/:name"
func GetService(c *gin.Context) {
	// 尝试解析请求里面的name
	name := c.Param("name")
	// log
	logStr := "GetSerive: name = " + name
	k8log.InfoLog("APIServer", logStr)

	// 如果解析成功，返回对应的Service信息
	if name != "" {
		res, err := etcdclient.EtcdStore.PrefixGet(serverconfig.EtcdServicePath + name)
		if err != nil {
			c.JSON(400, gin.H{
				"error": "get service failed " + err.Error(),
			})
			return
		}
		// 没找到
		if len(res) == 0 {
			c.JSON(404, gin.H{
				"error": "get service err, not find service",
			})
			return
		}

		// 处理res，如果发现有多个Service，返回错误
		if len(res) != 1 {
			c.JSON(500, gin.H{
				"error": "get service err, find more than one service",
			})
			return
		}
		// 遍历res，返回对应的Service信息
		targetService := res[0].Value
		c.JSON(200, gin.H{
			"data": targetService,
		})
		return
	} else {
		c.JSON(404, gin.H{
			"error": "name is empty",
		})
		return
	}
}

// 获取所有Service信息
func GetServices(c *gin.Context) {
	res, err := etcdclient.EtcdStore.PrefixGet(serverconfig.EtcdServicePath)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "get services failed " + err.Error(),
		})
		return
	}
	// 遍历res，返回对应的Service信息
	var services []string
	for _, service := range res {
		services = append(services, service.Value)
	}
	c.JSON(200, gin.H{
		"data": services,
	})
}

// 删除Service信息
func DeleteService(c *gin.Context) {
	// 尝试解析请求里面的name
	name := c.Params.ByName("name")
	// 如果解析成功，删除对应的Service信息
	if name != "" {
		// log
		logStr := "DeleteService: name = " + name
		k8log.InfoLog("APIServer", logStr)

		err := etcdclient.EtcdStore.Del(serverconfig.EtcdServicePath + name)
		if err != nil {
			c.JSON(400, gin.H{
				"error": "delete service failed " + err.Error(),
			})
			return
		}
		c.JSON(204, gin.H{
			"message": "delete service success",
		})
		return
	} else {
		c.JSON(404, gin.H{
			"error": "name is empty",
		})
		return
	}
}

// 更新Service信息
func UpdateService(c *gin.Context) {

}
