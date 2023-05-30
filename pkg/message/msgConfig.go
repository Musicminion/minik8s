package message

import (
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
)

// 这个文件是消息组件的配置文件，包括了消息组件的配置和消息组件的一些常量

// 可能使用到的ContentType
const ContentTypeJson = "application/json"
const ContentTypeText = "text/plain"

// 消息组件的配置
type MsgConfig struct {
	// RabbitMQ服务器地址
	User     string
	Password string
	Host     string
	Port     int
	// 虚拟Host
	VHost string
	// 最大重连次数 int
	MaxReconnect int
	// 重连间隔时间 s秒
	ReconnectInterval int
}

// 默认的配置是连接本地的RabbitMQ服务器，使用Guest账号
func DefaultMsgConfig() *MsgConfig {
	config := MsgConfig{
		User:     "admin",
		Password: "123456",
		Host:     config.GetMasterIP(),
		Port:     5672,
		VHost:    "/",
	}
	return &config
}

// 消息队列
const (
	NodeScheduleQueue = "nodeSchedule"
	EndpointUpdateQueue = "endpointUpdate"
	PodUpdateQueue = "podUpdate"
	ServiceUpdateQueue = "serviceUpdate"
	JobUpdateQueue = "jobUpdate"
	DnsUpdateQueue = "dnsUpdate"
	HostUpdateQueue = "hostUpdate"
)

// 根据node来路由消息到不同的队列
func PodUpdateWithNode(node string) string {
	return PodUpdateQueue + "-" + node
}

func HostUpdateWithNode(node string) string {
	k8log.DebugLog("message", "HostUpdateWithNode: "+HostUpdateQueue+"-"+node)
	return HostUpdateQueue + "-" + node
}

func ServiceUpdateWithNode(node string) string {
	return ServiceUpdateQueue + "-" + node
}

// K8s消息交换机名字
const DirectK8sExchange = "DirectK8sExchange"
const ServiecUpdateExchange = "ServiecUpdateExchange"
const HostUpdateExchange = "HostUpdateExchange"

// direct交换机
var DirectExchange = []string{
	DirectK8sExchange,
}

// fanout交换机
var FanoutExchange = []string{
	ServiecUpdateExchange,
	HostUpdateExchange,
}

// direct类型的队列
var DirectExchangeQueues = []string{
	NodeScheduleQueue,
	EndpointUpdateQueue,
	PodUpdateQueue,
	JobUpdateQueue,
	DnsUpdateQueue,
}

// Fanout类型的队列
var FanoutExchangeQueues = []string{
	ServiceUpdateQueue,
	HostUpdateQueue,
}

var queueToExchange = map[string]string{
	NodeScheduleQueue:   DirectK8sExchange,
	EndpointUpdateQueue: DirectK8sExchange,
	PodUpdateQueue:      DirectK8sExchange,
	JobUpdateQueue:      DirectK8sExchange,
	DnsUpdateQueue:      DirectK8sExchange,

	ServiceUpdateQueue: ServiecUpdateExchange,
	HostUpdateQueue:    HostUpdateExchange,
}
