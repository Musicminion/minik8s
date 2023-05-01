package scheduler

import (
	"encoding/json"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/listwatcher"
	"miniK8s/pkg/message"

	"github.com/streadway/amqp"
	"sync"
	"time"
	"math/rand"
)

type SchedulePolicy string
var globalCount int
var lock sync.Mutex
const (
	RoundRobin SchedulePolicy = "RoundRobin" // 轮询调度策略
	Random     SchedulePolicy = "Random"     // Random调度,产生一个随机数
	LeastPod   SchedulePolicy = "LeastPod"   // LeastPod数量调度,选择Pod数量最少的节点
	LeastCpu   SchedulePolicy = "LeastCpu"   // LeastCpu调度,选择CPU最少的节点
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
func schRoundRobin(nodes []apiObject.NodeStore) string {
	lock.Lock()
	defer lock.Unlock()
	cnt := len(nodes)
	if cnt == 0 {
		return ""
	}
	idx := globalCount % cnt
	globalCount++
	return nodes[idx].GetName()
}
func schRandom(nodes []apiObject.NodeStore) string {
	lock.Lock()
	defer lock.Unlock()
	cnt := len(nodes)
	if cnt == 0 {
		return ""
	}
	seconds := time.Now().Unix() //获取当前日期和时间的整数形式
	rand.Seed(seconds)           //播种随机生成器
	idx := rand.Intn(cnt) //生成一个介于0和cnt-1之间的整数
	return nodes[idx].GetName()
}
func schLeastPod(nodes []apiObject.NodeStore) string {
	lock.Lock()
	defer lock.Unlock()
	cnt := len(nodes)
	if cnt == 0 {
		return ""
	}
	//to do
	return ""
}
func schLeastCpu(nodes []apiObject.NodeStore) string {
	lock.Lock()
	defer lock.Unlock()
	cnt := len(nodes)
	if cnt == 0 {
		return ""
	}
	//to do
	return ""
}
func schLeastMem(nodes []apiObject.NodeStore) string {
	lock.Lock()
	defer lock.Unlock()
	cnt := len(nodes)
	if cnt == 0 {
		return ""
	}
	//to do
	return ""
}
/*********************************************************************/
/*********************************************************************/
// 从所有的节点里面选择一个节点
func (sch *Scheduler) ChooseFromNodes(nodes []apiObject.NodeStore) string {
	if len(nodes) == 0 {
        return ""
    }
    switch sch.polocy {
	case RoundRobin:
		return schRoundRobin(nodes)
	case Random:
		return schRandom(nodes)
	case LeastPod:
		return schLeastPod(nodes)
	case LeastCpu:
		return schLeastCpu(nodes)
	case LeastMem:
		return schLeastMem(nodes)
	default:
	}	
	// TODO
	return "ubuntu"
}

// 处理调度请求的消息
func (sch *Scheduler) RequestSchedule(parsedMsg *message.Message) {
	// TODO
	k8log.DebugLog("scheduler", "收到调度请求消息"+parsedMsg.Content)

	nodes, err := sch.GetAllNodes()

	if err != nil {
		k8log.ErrorLog("scheduler", "获取所有节点失败"+err.Error())
	}

	scheduledNode := sch.ChooseFromNodes(nodes)

	if scheduledNode == "" {
		k8log.ErrorLog("scheduler", "没有可用的节点")
		return
	}

	respMessage := message.Message{
		Type:         message.ScheduleResult,
		Content:      scheduledNode,
		ResourceURI:  parsedMsg.ResourceURI,
		ResourceName: parsedMsg.ResourceName,
	}

	// JSOn序列化
	result, err := json.Marshal(respMessage)
	if err != nil {
		k8log.ErrorLog("scheduler", "序列化消息失败")
	}

	sch.publisher.Publish("apiServer", message.ContentTypeJson, result)
}

// 调度器的消息处理函数,分发给不同的消息处理函数
func (sch *Scheduler) MsgHandler(msg amqp.Delivery) {
	parsedMsg, err := message.ParseJsonMessageFromBytes(msg.Body)
	if err != nil {
		k8log.ErrorLog("scheduler", "消息格式错误,无法转换为Message")
	}

	switch parsedMsg.Type {
	case message.RequestSchedule:
		// TODO
		sch.RequestSchedule(parsedMsg)
	default:

		// TODO
	}

}

// 启动调度器
func (sch *Scheduler) Run() {
	sch.lw.WatchQueue_Block("scheduler", sch.MsgHandler, make(chan struct{}))
}
