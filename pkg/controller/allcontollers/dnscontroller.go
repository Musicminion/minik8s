package allcontollers

import (
	"encoding/json"
	"miniK8s/pkg/apiObject"
	msgutil "miniK8s/pkg/apiserver/msgUtil"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/listwatcher"
	"miniK8s/pkg/message"
	"miniK8s/util/file"
	netrequest "miniK8s/util/netRequest"
	"miniK8s/util/nginx"
	"miniK8s/util/stringutil"
	"net/http"

	"github.com/streadway/amqp"
	"gopkg.in/yaml.v2"
)

const NginxPodYamlPath = "../../../util/nginx/yaml/dns-nginx-pod.yaml"
const NginxServiceYamlPath = "../../../util/nginx/yaml/dns-nginx-service.yaml"

type DnsController interface {
	Run()
}

type dnsController struct {
	lw *listwatcher.Listwatcher
}

func NewDnsController() (DnsController, error) {
	lwConfig := listwatcher.DefaultListwatcherConfig()
	newlw, err := listwatcher.NewListWatcher(lwConfig)

	if err != nil {
		return nil, err
	}

	return &dnsController{
		lw: newlw,
	}, nil
}

func (dc *dnsController) DnsCreateHandler(parsedMsg *message.Message) {
	// TODO
	dns := &apiObject.Dns{}
	err := json.Unmarshal([]byte(parsedMsg.Content), dns)
	if err != nil {
		k8log.ErrorLog("Job-Controller", "HandleServiceUpdate: failed to unmarshal")
		return
	}

	// 为nginx创建conf文件
	err = dc.CreateNginxConf(dns)
	if err != nil {
		k8log.ErrorLog("Job-Controller", "HandleServiceUpdate: failed to create nginx conf")
		return
	}

}

func (dc *dnsController) DnsUpdateHandler(parsedMsg *message.Message) {

}

func (dc *dnsController) DnsDeleteHandler(parsedMsg *message.Message) {

}

func (dc *dnsController) MsgHandler(msg amqp.Delivery) {
	k8log.WarnLog("Job-Controller", "收到消息"+string(msg.Body))

	parsedMsg, err := message.ParseJsonMessageFromBytes(msg.Body)
	if err != nil {
		k8log.ErrorLog("Job-Controller", "消息格式错误,无法转换为Message")
	}

	switch parsedMsg.Type {
	case message.UPDATE:
		dc.DnsCreateHandler(parsedMsg)
	case message.DELETE:
		dc.DnsDeleteHandler(parsedMsg)
	}
}

func (dc *dnsController) CreateNginxConf(dns *apiObject.Dns) error {
	conf := nginx.FormatConf(*dns)
	err := nginx.WriteConf(*dns, conf)
	if err != nil {
		k8log.ErrorLog("Job-Controller", "CreateNginxConf: failed to write conf"+err.Error())
		return err
	}
	return nil
}

func (dc *dnsController) CreateNginxPod() {
	path := NginxPodYamlPath
	fileContent, err := file.ReadFile(path)
	if err != nil {
		k8log.ErrorLog("Job-Controller", "Run: failed to read file"+err.Error())
		return
	}

	// 将文件内容转换为Pod对象
	// 通过调用gin引擎的ServeHTTP方法，可以模拟一个http请求，从而测试AddPod方法。
	nginxPod := &apiObject.Pod{}
	err = yaml.Unmarshal(fileContent, nginxPod)
	if err != nil {
		k8log.ErrorLog("Job-Controller", "Run: failed to unmarshal"+err.Error())
		return
	}

	URL := stringutil.Replace(config.PodsURL, config.URL_PARAM_NAMESPACE_PART, nginxPod.GetPodNamespace())
	URL = config.API_Server_URL_Prefix + URL
	k8log.DebugLog("Job-Controller", "Run: URL is "+URL)
	code, _, err := netrequest.PostRequestByTarget(URL, nginxPod)
	if err != nil {
		k8log.ErrorLog("Job-Controller", "Run: failed to post request"+err.Error())
		return
	}
	if code != http.StatusCreated {
		k8log.ErrorLog("Job-Controller", "Run: failed to create pod")
		return
	}

	k8log.InfoLog("Job-Controller", "HandleServiceUpdate: success to create nginx pod")
}

func (dc *dnsController) CreateNginxService() {
	path := NginxServiceYamlPath
	fileContent, err := file.ReadFile(path)
	if err != nil {
		k8log.ErrorLog("Job-Controller", "Run: failed to read file"+err.Error())
		return
	}

	// 将文件内容转换为Pod对象
	// 通过调用gin引擎的ServeHTTP方法，可以模拟一个http请求，从而测试AddPod方法。
	nginxService := &apiObject.Service{}
	err = yaml.Unmarshal(fileContent, nginxService)
	if err != nil {
		k8log.ErrorLog("Job-Controller", "Run: failed to unmarshal"+err.Error())
		return
	}

	URL := stringutil.Replace(config.ServiceURL, config.URL_PARAM_NAMESPACE_PART, nginxService.GetNamespace())
	URL = config.API_Server_URL_Prefix + URL
	k8log.DebugLog("Job-Controller", "Run: URL is "+URL)
	code, _, err := netrequest.PostRequestByTarget(URL, nginxService)
	if err != nil {
		k8log.ErrorLog("Job-Controller", "Run: failed to post request"+err.Error())
		return
	}
	if code != http.StatusCreated {
		k8log.ErrorLog("Job-Controller", "Run: failed to create service")
		return
	}

	k8log.InfoLog("Job-Controller", "HandleServiceUpdate: success to create nginx service")
}

func (dc *dnsController) Run() {
	// 在每个node上创建一个nginx pod
	// 1. 创建nginx pod
	// 2. 创建nginx service
	// 3. 创建nginx dns
	dc.CreateNginxPod()
	dc.CreateNginxService()


	dc.lw.WatchQueue_Block(msgutil.DnsUpdate, dc.MsgHandler, make(chan struct{}))
}
