package scheduler

import (
	"encoding/json"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/listwatcher"
	"miniK8s/pkg/message"
	netrequest "miniK8s/util/netRequest"
	"miniK8s/util/stringutil"
	"net/http"
	"strconv"

	"math/rand"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type SchedulePolicy string

var globalCount int
var lock sync.Mutex

const (
	RoundRobin SchedulePolicy = "RoundRobin" // 轮询调度策略
	Random     SchedulePolicy = "Random"     // Random调度,产生一个随机数
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

	newPublisher, err := message.NewPublisher(message.DefaultMsgConfig())
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
	k8log.InfoLog("scheduler start with config: %s", string(schedulerConfig.Policy))
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
	k8log.DebugLog("Scheduler", "RoundRobin调度策略选择节点"+nodes[idx].GetName())
	return nodes[idx].GetName()
}
func schRandom(nodes []apiObject.NodeStore) string {
	lock.Lock()
	defer lock.Unlock()
	cnt := len(nodes)
	if cnt == 0 {
		return ""
	}
	// seconds := time.Now().Unix() //获取当前日期和时间的整数形式
	r := rand.New(rand.NewSource(time.Now().Unix()))
	idx := r.Intn(cnt)
	return nodes[idx].GetName()
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
	default:
	}
	// TODO:
	return ""
}

// 处理调度请求的消息
func (sch *Scheduler) RequestSchedule(parsedMsg *message.Message) {
	// TODO
	k8log.DebugLog("Scheduler", "收到调度请求消息"+parsedMsg.Content)

	allNodes, err := sch.GetAllNodes()

	if err != nil {
		k8log.ErrorLog("Scheduler", "获取所有节点失败"+err.Error())
	}

	// 调度的时候筛选存活的节点
	nodes := make([]apiObject.NodeStore, 0)
	for _, node := range allNodes {
		if node.Status.Condition == apiObject.Ready {
			nodes = append(nodes, node)
		}
	}

	// 反序列化pod
	podStore := &apiObject.PodStore{}
	err = json.Unmarshal([]byte(parsedMsg.Content), &podStore)
	if err != nil {
		k8log.ErrorLog("Scheduler", "反序列化pod失败")
		return
	}

	var scheduledNode string

	// 如果在pod中指定了node
	if podStore.Spec.NodeName != "" {
		// 检查node是否存在
		for _, node := range nodes {
			if node.GetName() == podStore.Spec.NodeName {
				scheduledNode = podStore.Spec.NodeName
			}
		}
	}

	// 如果未指定node或者指定的node无效，则选择一个节点
	if scheduledNode == "" {
		scheduledNode = sch.ChooseFromNodes(nodes)
	}

	if scheduledNode == "" {
		k8log.ErrorLog("Scheduler", "没有可用的节点")
		return
	}

	// 为pod添加node信息
	podStore.Spec.NodeName = scheduledNode

	// 更新Apiserver中的Pod信息
	URL := stringutil.Replace(config.PodSpecURL, config.URL_PARAM_NAMESPACE_PART, podStore.GetPodNamespace())
	URL = stringutil.Replace(URL, config.URL_PARAM_NAME_PART, podStore.GetPodName())
	URL = config.GetAPIServerURLPrefix() + URL

	code, _, err := netrequest.PutRequestByTarget(URL, podStore)
	if err != nil {
		k8log.ErrorLog("Scheduler", "更新Pod信息失败"+err.Error())
		return
	}
	if code != http.StatusOK {
		k8log.ErrorLog("Scheduler", "更新Pod信息失败,code: "+strconv.Itoa(code))
		return
	}

	podUpdate := &entity.PodUpdate{
		Action:    message.CREATE,
		PodTarget: *podStore,
		Node:      scheduledNode,
	}
	message.PublishUpdatePod(podUpdate)
}

// 调度器的消息处理函数,分发给不同的消息处理函数
func (sch *Scheduler) MsgHandler(msg amqp.Delivery) {
	k8log.DebugLog("Scheduler", "收到消息"+string(msg.Body))
	parsedMsg, err := message.ParseJsonMessageFromBytes(msg.Body)
	if err != nil {
		k8log.ErrorLog("Scheduler", "消息格式错误,无法转换为Message")
	}

	// 处理调度请求的消息
	sch.RequestSchedule(parsedMsg)
}

func (sch *Scheduler) Run() {
	// TODO
	// 监听队列
	for {
		// 监听队列
		sch.lw.WatchQueue_Block(message.NodeScheduleQueue, sch.MsgHandler, make(chan struct{}))
	}
}
