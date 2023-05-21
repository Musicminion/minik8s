package allcontollers

import (
	"encoding/json"
	"miniK8s/pkg/apiObject"
	msgutil "miniK8s/pkg/apiserver/msgUtil"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/listwatcher"
	"miniK8s/pkg/message"

	"github.com/streadway/amqp"
)

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



func (dc *dnsController) CreateNginxConf(dns *apiObject.Dns) error{
	// TODO

	return nil
}
	


func (dc *dnsController) Run() {
	// 在每个node上创建一个nginx pod
	// 1. 创建nginx pod
	// 2. 创建nginx service
	// 3. 创建nginx dns

	

	dc.lw.WatchQueue_Block(msgutil.DnsUpdate, dc.MsgHandler, make(chan struct{}))
}
