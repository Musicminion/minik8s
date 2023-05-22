package handlers

import (
	"encoding/json"
	"miniK8s/pkg/apiObject"
	etcdclient "miniK8s/pkg/apiserver/app/etcdclient"
	msgutil "miniK8s/pkg/apiserver/msgUtil"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/config"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/message"
	"miniK8s/util/uuid"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
)

// 创建Dns
// "/apis/v1/namespaces/:namespace/jobs"
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

	// 将dns存储到etcd中
	dns.Metadata.UUID = uuid.NewUUID()

	dnsJson, err := json.Marshal(dns)

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

	// 返回
	c.JSON(http.StatusCreated, gin.H{
		"message": "create dns success",
	})

	dnsUpdate := entity.DnsUpdate{
		Action:    message.CREATE,
		DnsTarget: dns,
	}
	msgutil.PublishUpdateDns(&dnsUpdate)
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

	// 删除dns
	key := path.Join(serverconfig.EtcdDnsPath, namespace, name)
	err := etcdclient.EtcdStore.Del(key)
	if err != nil {
		k8log.ErrorLog("APIServer", "DeleteDns, err is "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 返回
	c.JSON(http.StatusNoContent, gin.H{
		"message": "delete job success",
	})

	k8log.InfoLog("APIServer", "delete dns success")
	dnsUpdate := entity.DnsUpdate{
		Action:    message.DELETE,
		DnsTarget: apiObject.Dns{
			Basic: apiObject.Basic{
				Metadata: apiObject.Metadata{
					Name:      name,
					Namespace: namespace,
				},
			},
		},
	}
	msgutil.PublishUpdateDns(&dnsUpdate)
}

// 获取Dns
func GetDns(c *gin.Context) {
	k8log.InfoLog("APIServer", "GetDns")
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