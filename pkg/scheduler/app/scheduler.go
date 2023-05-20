package scheduler

import (
	"encoding/json"
	"miniK8s/pkg/apiObject"
	msgutil "miniK8s/pkg/apiserver/msgUtil"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/listwatcher"
	"miniK8s/pkg/message"

	"github.com/streadway/amqp"
)

type SchedulePolicy string

const (
	RoundRobin SchedulePolicy = "RoundRobin" // 轮询调度策略
	Random     SchedulePolicy = "Random"     // Random调度,产生一个随机数
	LeastPod   SchedulePolicy = "LeastPod"   // LeastPod数量调度,选择Pod数量最少的节点
	LeastCpu   SchedulePolicy = "LeastPod"   // LeastCpu调度,选择CPU最少的节点
	LeastMem   SchedulePolicy = "LeastMem"   // LeastMem调度,选择内存最少的节点
)

type Scheduler struct {
	// listwatcher
	lw *listwatcher.Listwatcher
	// Publisher
	publisher *message.Publisher
	// 调度策略
	polocy SchedulePolicy
	// apiServer的地址
	apiServerHost string
	// apiServer的端口
	apiServerPort int
}

// 创建一个调度器
func NewScheduler() (*Scheduler, error) {
	newlistwatcher, err := listwatcher.NewListWatcher(listwatcher.DefaultListwatcherConfig())
	if err != nil {
		return nil, err
	}
	schedulerConfig := DefaultSchedulerConfig()

	messageConfig := message.MsgConfig{
		User:     "guest",
		Password: "guest",
		Host:     "localhost",
		Port:     5672,
		VHost:    "/",
	}
	newPublisher, err := message.NewPublisher(&messageConfig)
	if err != nil {
		return nil, err
	}
	scheduler := &Scheduler{
		lw:            newlistwatcher,
		polocy:        schedulerConfig.Policy,
		apiServerHost: schedulerConfig.ApiServerHost,
		apiServerPort: schedulerConfig.ApiServerPort,
		publisher:     newPublisher,
	}
	return scheduler, nil
}

/* 根据具体的调度策略选择一个节点 */
/*********************************************************************/
/*********************************************************************/

/*********************************************************************/
/*********************************************************************/
// 从所有的节点里面选择一个节点
func (sch *Scheduler) ChooseFromNodes(nodes []apiObject.NodeStore) string {
	// TODO
	return "ubuntu"
}

// 处理调度请求的消息
func (sch *Scheduler) RequestSchedule(parsedMsg *message.Message) {
	// TODO
	k8log.DebugLog("Scheduler", "收到调度请求消息"+parsedMsg.Content)

	nodes, err := sch.GetAllNodes()

	if err != nil {
		k8log.ErrorLog("Scheduler", "获取所有节点失败"+err.Error())
	}

	scheduledNode := sch.ChooseFromNodes(nodes)

	if scheduledNode == "" {
		k8log.ErrorLog("Scheduler", "没有可用的节点")
		return
	}

	// respMessage := message.Message{
	// 	Type:         message.ScheduleResult,
	// 	Content:      scheduledNode,
	// 	ResourceURI:  parsedMsg.ResourceURI,
	// 	ResourceName: parsedMsg.ResourceName,
	// }

	// JSOn序列化
	// result, err := json.Marshal(respMessage)
	// if err != nil {
	// 	k8log.ErrorLog("Scheduler", "序列化消息失败")
	// }

	podStore := &apiObject.PodStore{}
	err = json.Unmarshal([]byte(parsedMsg.Content), &podStore)
	if err != nil {
		k8log.ErrorLog("Scheduler", "反序列化pod失败")
		return
	}

	// 为pod添加node信息
	podStore.Spec.NodeName = scheduledNode

	// TODO: 将podUpdate发送给对应的Node
	podUpdate := &entity.PodUpdate{
		Action:    message.CREATE,
		PodTarget: *podStore,
		Node:      scheduledNode,
	}
	msgutil.PublishUpdatePod(podUpdate)
	// sch.publisher.Publish("apiServer", message.ContentTypeJson, result)
}

// 调度器的消息处理函数,分发给不同的消息处理函数
func (sch *Scheduler) MsgHandler(msg amqp.Delivery) {
	k8log.DebugLog("Scheduler", "收到消息"+string(msg.Body))
	parsedMsg, err := message.ParseJsonMessageFromBytes(msg.Body)
	if err != nil {
		k8log.ErrorLog("Scheduler", "消息格式错误,无法转换为Message")
	}

	switch parsedMsg.Type {
	case message.RequestSchedule:
		// TODO
		sch.RequestSchedule(parsedMsg)
	default:
		// TODO
	}

}

func (sch *Scheduler) Run() {
	// TODO
	// 监听队列
	for {
		// 监听队列
		sch.lw.WatchQueue_Block(msgutil.NodeSchedule, sch.MsgHandler, make(chan struct{}))
	}
}
