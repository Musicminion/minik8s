package allcontollers

import (
	"encoding/json"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/listwatcher"
	"miniK8s/pkg/message"
	"miniK8s/util/file"
	netrequest "miniK8s/util/netRequest"
	"miniK8s/util/nginx"
	"miniK8s/util/stringutil"
	"net/http"
	"os"

	"github.com/streadway/amqp"
	"gopkg.in/yaml.v2"
)

var NginxPodYamlPath = os.Getenv("MINIK8S_PATH") + "util/nginx/yaml/dns-nginx-pod.yaml"
var NginxServiceYamlPath = os.Getenv("MINIK8S_PATH") + "util/nginx/yaml/dns-nginx-service.yaml"
var NginxDnsYamlPath = os.Getenv("MINIK8S_PATH") + "util/nginx/yaml/dns-nginx-dns.yaml"

type DnsController interface {
	Run()
}

type dnsController struct {
	lw           *listwatcher.Listwatcher
	hostList     []string // 通过dns创建的host列表，这些host将被解析为nginx的service ip
	nginxSvcName string   // nginx service的名称
	nginxSvcIp   string   // nginx service的ip
}

func NewDnsController() (DnsController, error) {
	lwConfig := listwatcher.DefaultListwatcherConfig()
	newlw, err := listwatcher.NewListWatcher(lwConfig)

	if err != nil {
		return nil, err
	}

	return &dnsController{
		lw:       newlw,
		hostList: make([]string, 0),
	}, nil
}

func (dc *dnsController) DnsCreateHandler(parsedMsg *message.Message) {
	dnsUpdate := &entity.DnsUpdate{}
	err := json.Unmarshal([]byte(parsedMsg.Content), dnsUpdate)
	if err != nil {
		k8log.ErrorLog("Dns-Controller", "failed to unmarshal")
		return
	}

	dnsStore := dnsUpdate.DnsTarget

	if dnsStore.Spec.Host == "" {
		k8log.ErrorLog("Dns-Controller", "host is empty")
		return
	}

	if dnsStore.Metadata.Namespace == "" {
		dnsStore.Metadata.Namespace = config.DefaultNamespace
	}

	// 为nginx创建conf文件
	nginxConfig := nginx.FormatConf(*dnsStore.ToDns())

	// 添加/etc/hosts
	k8log.DebugLog("Dns-Controller", "DnsCreateHandler: newhostEntry is "+dc.nginxSvcIp+" "+dnsStore.Spec.Host)
	newHostEntry := dc.nginxSvcIp + " " + dnsStore.Spec.Host
	dc.hostList = append(dc.hostList, newHostEntry)

	// 创建hostUpdate消息
	hostUpdate := &entity.HostUpdate{
		Action:    message.CREATE,
		DnsTarget: dnsStore,
		DnsConfig: nginxConfig,
		HostList:  dc.hostList,
	}

	// TODO: 通知所有的节点进行hosts的修改
	k8log.DebugLog("Dns-Controller", "DnsCreateHandler: publish hostUpdate")
	message.PubelishUpdateHost(hostUpdate)
}

func (dc *dnsController) DnsDeleteHandler(parsedMsg *message.Message) {
	dnsUpdate := &entity.DnsUpdate{}
	err := json.Unmarshal([]byte(parsedMsg.Content), dnsUpdate)
	if err != nil {
		k8log.ErrorLog("Dns-Controller", "failed to unmarshal")
		return
	}

	dnsStore := dnsUpdate.DnsTarget

	if dnsStore.Spec.Host == "" {
		k8log.ErrorLog("Dns-Controller", "host is empty")
		return
	}

	if dnsStore.Metadata.Namespace == "" {
		dnsStore.Metadata.Namespace = config.DefaultNamespace
	}

	// 删除nginx conf文件
	err = dc.DeleteNginxConf(dnsStore.ToDns())
	if err != nil {
		k8log.ErrorLog("Dns-Controller", "DnsDeleteHandler: failed to delete nginx conf")
		return
	}

	// 删除/etc/hosts
	k8log.DebugLog("Dns-Controller", "DnsDeleteHandler: newhostEntry is "+dc.nginxSvcIp+" "+dnsStore.Spec.Host)
	deleteHostEntry := dc.nginxSvcIp + " " + dnsStore.Spec.Host
	// 删除hostList中的host
	for i, hostEntry := range dc.hostList {
		if hostEntry == deleteHostEntry {
			dc.hostList = append(dc.hostList[:i], dc.hostList[i+1:]...)
			break
		}
	}
	// 创建hostUpdate消息
	hostUpdate := &entity.HostUpdate{
		Action:    message.DELETE,
		DnsTarget: dnsStore,
		HostList:  dc.hostList,
	}

	// TODO: 通知所有的节点进行hosts的修改
	k8log.DebugLog("Dns-Controller", "DnsCreateHandler: publish hostUpdate")
	message.PubelishUpdateHost(hostUpdate)
}

func (dc *dnsController) DnsUpdatehandler(msg amqp.Delivery) {
	k8log.WarnLog("Dns-Controller", "收到消息"+string(msg.Body))

	parsedMsg, err := message.ParseJsonMessageFromBytes(msg.Body)
	if err != nil {
		k8log.ErrorLog("Dns-Controller", "消息格式错误,无法转换为Message")
	}

	switch parsedMsg.Type {
	case message.CREATE:
		dc.DnsCreateHandler(parsedMsg)
	case message.DELETE:
		dc.DnsDeleteHandler(parsedMsg)
	case message.UPDATE:
		dc.DnsDeleteHandler(parsedMsg)
		dc.DnsCreateHandler(parsedMsg)
	}
}

func (dc *dnsController) DeleteNginxConf(dns *apiObject.Dns) error {
	// 删除nginx conf文件
	err := nginx.DeleteConf(*dns)
	if err != nil {
		k8log.ErrorLog("Dns-Controller", "DeleteNginxConf: failed to delete conf"+err.Error())
		return err
	}
	return nil
}

func (dc *dnsController) CreateNginxService() {
	path := NginxServiceYamlPath
	fileContent, err := file.ReadFile(path)
	if err != nil {
		k8log.ErrorLog("Dns-Controller", "Run: failed to read file"+err.Error())
		return
	}

	// 将文件内容转换为Service对象
	nginxService := &apiObject.Service{}
	err = yaml.Unmarshal(fileContent, nginxService)
	if err != nil {
		k8log.ErrorLog("Dns-Controller", "Run: failed to unmarshal"+err.Error())
		return
	}

	URL := stringutil.Replace(config.ServiceURL, config.URL_PARAM_NAMESPACE_PART, nginxService.GetObjectNamespace())
	URL = config.GetAPIServerURLPrefix() + URL
	k8log.DebugLog("Dns-Controller", "Run: URL is "+URL)
	code, _, err := netrequest.PostRequestByTarget(URL, nginxService)
	if err != nil {
		k8log.ErrorLog("Dns-Controller", "Run: failed to post request"+err.Error())
		return
	}
	if code != http.StatusCreated {
		k8log.ErrorLog("Dns-Controller", "Run: failed to create service")
		return
	}

	// 更新nginx service的名称
	dc.nginxSvcName = nginxService.GetObjectName()
	k8log.InfoLog("Dns-Controller", "HandleServiceUpdate: success to create nginx service")
}

func (dc *dnsController) UpdateNginxSvcIP() {
	// 获取nginx service的ip
	// 通过api server获取nginx service的ip
	nginxSvc := &apiObject.Service{}
	URL := stringutil.Replace(config.ServiceSpecURL, config.URL_PARAM_NAMESPACE_PART, config.DefaultNamespace)
	URL = stringutil.Replace(URL, config.URL_PARAM_NAME_PART, dc.nginxSvcName)
	URL = config.GetAPIServerURLPrefix() + URL
	k8log.DebugLog("Dns-Controller", "Run: URL is "+URL)
	code, err := netrequest.GetRequestByTarget(URL, nginxSvc, "data")
	if err != nil {
		k8log.ErrorLog("Dns-Controller", "UpdateNginxSvcIP: failed to get nginx service"+err.Error())
		return
	}
	if code != http.StatusOK {
		k8log.ErrorLog("Dns-Controller", "UpdateNginxSvcIP: failed to get nginx service")
		return
	}

	// 更新nginx service的ip
	k8log.DebugLog("Dns-Controller", "UpdateNginxSvcIP: nginx service ip is "+nginxSvc.Spec.ClusterIP)
	dc.nginxSvcIp = nginxSvc.Spec.ClusterIP
}

func (dc *dnsController) CreateNginxDns() {
	path := NginxDnsYamlPath
	fileContent, err := file.ReadFile(path)
	if err != nil {
		k8log.ErrorLog("Dns-Controller", "Run: failed to read file"+err.Error())
		return
	}

	// 将文件内容转换为Dns对象
	nginxDns := &apiObject.Dns{}
	err = yaml.Unmarshal(fileContent, nginxDns)
	if err != nil {
		k8log.ErrorLog("Dns-Controller", "Run: failed to unmarshal"+err.Error())
		return
	}

	URL := stringutil.Replace(config.DnsURL, config.URL_PARAM_NAMESPACE_PART, nginxDns.GetObjectNamespace())
	URL = config.GetAPIServerURLPrefix() + URL
	k8log.DebugLog("Dns-Controller", "Run: URL is "+URL)
	code, _, err := netrequest.PostRequestByTarget(URL, nginxDns)
	if err != nil {
		k8log.ErrorLog("Dns-Controller", "Run: failed to post request"+err.Error())
		return
	}
	if code != http.StatusCreated {
		k8log.ErrorLog("Dns-Controller", "Run: failed to create service")
		return
	}

	k8log.InfoLog("Dns-Controller", "HandleServiceUpdate: success to create nginx dns")
}

func (dc *dnsController) Run() {
	// 创建nginx service
	dc.CreateNginxService()
	// 更新nginxService的ip
	dc.UpdateNginxSvcIP()

	// 监听dns的更新
	dc.lw.WatchQueue_Block(message.DnsUpdateQueue, dc.DnsUpdatehandler, make(chan struct{}))
}
