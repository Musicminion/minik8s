package handlers

import (
	"encoding/json"
	"miniK8s/pkg/apiObject"
	etcdclient "miniK8s/pkg/apiserver/app/etcdclient"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/config"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/message"
	"miniK8s/util/stringutil"
	"miniK8s/util/uuid"
	"net/http"
	"path"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 创建Dns
// "/apis/v1/namespaces/:namespace/dns"
func AddDns(c *gin.Context) {
	k8log.InfoLog("APIServer", "AddDns")
	var dns apiObject.Dns
	// 从请求体中读取数据
	if err := c.ShouldBindJSON(&dns); err != nil {
		k8log.ErrorLog("APIServer", "AddDns, err is "+err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		k8log.ErrorLog("APIServer", "parse dns err, "+err.Error())
		return
	}

	// 将dns转换为字符串
	dnsString, _ := json.Marshal(dns)

	k8log.DebugLog("APIServer", "dns is "+string(dnsString))

	// 检查dns的合法性
	if dns.Metadata.Name == "" {
		k8log.ErrorLog("APIServer", "dns name is empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "dns name is empty",
		})
		return
	}

	if dns.Metadata.Namespace == "" {
		dns.Metadata.Namespace = config.DefaultNamespace
	}

	// 判断dns是否存在
	key := path.Join(serverconfig.EtcdDnsPath, dns.Metadata.Namespace, dns.Metadata.Name)
	res, err := etcdclient.EtcdStore.Get(key)
	if err != nil {
		k8log.ErrorLog("APIServer", "AddDns, err is "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	if len(res) != 0 {
		k8log.ErrorLog("APIServer", "dns already exists")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "dns already exists",
		})
		return
	}

	// 根据dns的path的service的名字查找出service的ip并回填到dns
	k8log.DebugLog("APIServer", strconv.Itoa(len(dns.Spec.Paths)))
	for i, p := range dns.Spec.Paths {
		// 获取service
		k8log.DebugLog("APIServer", "p.SvcName is "+p.SvcName)
		if p.SvcName == "" {
			k8log.ErrorLog("APIServer", "AddDns, err is service name is empty")
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "service name is empty",
			})
			return
		}

		serviceKey := path.Join(serverconfig.EtcdServicePath, dns.GetObjectNamespace(), p.SvcName)
		serviceRes, err := etcdclient.EtcdStore.Get(serviceKey)
		if err != nil {
			k8log.ErrorLog("APIServer", "AddDns, err is "+err.Error())
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if len(serviceRes) != 1 {
			k8log.ErrorLog("APIServer", "service not exists")
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "service not exists or number is wrong",
			})
			return
		}
		service := &apiObject.ServiceStore{}
		err = json.Unmarshal([]byte(serviceRes[0].Value), service)
		if err != nil {
			k8log.ErrorLog("APIServer", "AddDns, err is "+err.Error())
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		dns.Spec.Paths[i].SvcIp = service.Spec.ClusterIP
	}

	dns.Metadata.UUID = uuid.NewUUID()

	// 将dns转换为dnsStore
	dnsStore := dns.ToDnsStore()

	dnsJson, err := json.Marshal(dnsStore)

	if err != nil {
		k8log.ErrorLog("APIServer", "AddDns, err is "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = etcdclient.EtcdStore.Put(key, dnsJson)

	if err != nil {
		k8log.ErrorLog("APIServer", "AddDns, err is "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	dnsUpdate := entity.DnsUpdate{
		Action:    message.CREATE,
		DnsTarget: *dnsStore,
	}
	message.PublishUpdateDns(&dnsUpdate)

	// 返回
	c.JSON(http.StatusCreated, gin.H{
		"message": "create dns success",
	})
}

// 删除Dns
// "/apis/v1/namespaces/:namespace/dns/:name"
func DeleteDns(c *gin.Context) {
	k8log.InfoLog("APIServer", "DeleteDns")
	// 从url中获取dns的名称和命名空间
	name := c.Params.ByName("name")
	namespace := c.Params.ByName("namespace")

	// 检查参数是否为空
	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace or name is empty",
		})
		return
	}

	// 从etcd获取dns，之后发送dnsUpdate需要
	key := path.Join(serverconfig.EtcdDnsPath, namespace, name)
	dnsLRs, err := etcdclient.EtcdStore.Get(key)
	if err != nil {
		k8log.ErrorLog("APIServer", "DeleteDns, err is "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	if len(dnsLRs) != 1 {
		k8log.ErrorLog("APIServer", "dns not exists")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "dns not exists",
		})
		return
	}

	dns := &apiObject.HpaStore{}
	err = json.Unmarshal([]byte(dnsLRs[0].Value), dns)
	if err != nil {
		k8log.ErrorLog("APIServer", "DeleteDns, err is "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 删除dns
	err = etcdclient.EtcdStore.Del(key)
	if err != nil {
		k8log.ErrorLog("APIServer", "DeleteDns, err is "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 发送dnsUpdate
	dnsUpdate := entity.DnsUpdate{
		Action: message.DELETE,
		DnsTarget: apiObject.HpaStore{
			Spec: apiObject.DnsSpec{
				Host: dns.Spec.Host,
			},
			Basic: apiObject.Basic{
				Metadata: apiObject.Metadata{
					Name:      name,
					Namespace: namespace,
				},
			},
		},
	}
	message.PublishUpdateDns(&dnsUpdate)

	k8log.InfoLog("APIServer", "delete dns success")
	// 返回
	c.JSON(http.StatusNoContent, gin.H{
		"message": "delete job success",
	})
}

// 获取单个Dns
// "/apis/v1/namespaces/:namespace/dns/:name"
func GetDns(c *gin.Context) {
	k8log.InfoLog("APIServer", "GetDns")
	// 从url中获取dns的名称和命名空间
	name := c.Params.ByName("name")
	namespace := c.Params.ByName("namespace")

	// 检查参数是否为空
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "name is empty",
		})
		return
	}

	if namespace == "" {
		namespace = config.DefaultNamespace
	}

	// 获取dns
	key := path.Join(serverconfig.EtcdDnsPath, namespace, name)
	res, err := etcdclient.EtcdStore.Get(key)
	if err != nil {
		k8log.ErrorLog("APIServer", "GetDns, err is "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if len(res) == 0 {
		k8log.ErrorLog("APIServer", "dns not exists")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "dns not exists",
		})
		return
	}

	if len(res) != 1 {
		k8log.ErrorLog("APIServer", "dns not exists")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "dns not exists",
		})
		return
	}

	// 返回
	c.JSON(http.StatusOK, gin.H{
		"data": string(res[0].Value),
	})

	k8log.DebugLog("APIServer", "dns : "+res[0].Value)

	k8log.InfoLog("APIServer", "get dns success")
}

// 获取所有Dns
// "/apis/v1/namespaces/:namespace/dns"
func GetDnsList(c *gin.Context) {
	k8log.InfoLog("APIServer", "GetDnsList")
	// 从url中获取dns的名称和命名空间
	namespace := c.Params.ByName("namespace")

	// 检查参数是否为空
	if namespace == "" {
		namespace = config.DefaultNamespace
	}

	// 获取dns
	key := path.Join(serverconfig.EtcdDnsPath, namespace)
	res, err := etcdclient.EtcdStore.PrefixGet(key)
	if err != nil {
		k8log.ErrorLog("APIServer", "GetDnsList, err is "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 遍历res，返回对应的Dns信息
	targetDnsString := make([]string, 0)
	for _, dns := range res {
		targetDnsString = append(targetDnsString, dns.Value)
	}

	// 返回
	c.JSON(http.StatusOK, gin.H{
		"data": stringutil.StringSliceToJsonArray(targetDnsString),
	})

	k8log.InfoLog("APIServer", "get dns list success")
}
