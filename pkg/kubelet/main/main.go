package main

import (
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/kubelet/kubeletconfig"
	"miniK8s/pkg/kubelet/status"
	"miniK8s/pkg/kubelet/worker"
	"miniK8s/pkg/listwatcher"
)

type Kubelet struct {
	config        *kubeletconfig.KubeletConfig
	lw            *listwatcher.Listwatcher
	workManager   worker.PodWorkerManager
	statusManager status.StatusManager
}

func NewKubelet(conf *kubeletconfig.KubeletConfig) (*Kubelet, error) {
	newlw, err := listwatcher.NewListWatcher(conf.LWConf)
	if err != nil {
		return nil, err
	}

	k := &Kubelet{
		config:        conf,
		lw:            newlw,
		workManager:   worker.NewPodWorkerManager(),
		statusManager: status.NewStatusManager(),
	}

	return k, nil
}

func (k *Kubelet) Run() {
	go k.statusManager.Run()

}

func main() {
	KubeleConfig := kubeletconfig.DefaultKubeletConfig()
	Kubelet, err := NewKubelet(KubeleConfig)
	if err != nil {
		k8log.FatalLog("Kublet", "NewKubelet failed, for "+err.Error())
		return
	}

	Kubelet.Run()
}

// type Kubelet struct {
// 	config *config.KubeletConfig
// }

// 初始化的时候给API Server注册当前的节点
// func registerNode() error {
// 	k8log.InfoLog("Kublet", "registerNode")

// 	hostName := host.GetHostName()
// 	if hostName == "" {
// 		k8log.FatalLog("Kublet", "registerNode failed, for hostName is empty")
// 		return nil
// 	}

// 	ip, err := host.GetHostIp()
// 	if err != nil {
// 		k8log.FatalLog("Kublet", "registerNode failed, for get host ip failed")
// 		return nil
// 	}

// 	if ip == "" {
// 		k8log.FatalLog("Kublet", "registerNode failed, for host ip is empty")
// 		return nil
// 	}

// 	// netrequest.GetRequest("")
// 	// 发送GET请求给API Server，获取当前的Node信息
// 	code, res, err := netrequest.GetRequest("http://127.0.0.1:8090/api/v1/nodes/" + hostName)

// 	if err != nil {
// 		k8log.ErrorLog("Kublet", "registerNode failed, for get node info failed")
// 		return nil
// 	}

// 	if code == 200 {
// 		// 如果获取成功，说明当前的Node已经注册过了，直接返回
// 		k8log.InfoLog("Kublet", "registerNode Already, for node has been registered")
// 		jsonBytes, err := json.Marshal(res)
// 		if err != nil {
// 			k8log.ErrorLog("Kublet", "registerNode failed, for marshal node info failed")
// 			return nil
// 		}
// 		// json.Unmarshal(jsonBytes, &jsonBytes)
// 		// jsonString, _ := json.Marshal(jsonBytes)
// 		k8log.DebugLog("Kublet", "registerNode resp: "+string(jsonBytes))

// 		return nil
// 	} else if code == 404 {
// 		registResult := false
// 		k8log.DebugLog("Kublet", "try to registerNode")
// 		// 如果注册失败持续尝试
// 		for !registResult {
// 			node := apiObject.Node{
// 				NodeBasic: apiObject.NodeBasic{
// 					APIVersion: "v1",
// 					Kind:       "Node",
// 					NodeMetadata: apiObject.NodeMetadata{
// 						Name: hostName,
// 						UUID: "",
// 					},
// 				},
// 				IP: ip,
// 			}
// 			code, res, err := netrequest.PostRequestByTarget("http://127.0.0.1:8090/api/v1/nodes/", node)
// 			if err != nil {
// 				k8log.ErrorLog("Kublet", "registerNode failed, for post node info failed")
// 				return nil
// 			}
// 			jsonBytes, err := json.Marshal(res)
// 			if err != nil {
// 				k8log.ErrorLog("Kublet", "registerNode failed, for marshal node info failed")
// 				return nil
// 			}
// 			k8log.InfoLog("Kublet", "registerNode resp: "+string(jsonBytes))
// 			if code == 201 {
// 				k8log.InfoLog("Kublet", "registerNode success")
// 				return nil
// 			} else {
// 				k8log.ErrorLog("Kublet", "registerNode failed, for post node info failed")
// 				return nil
// 			}
// 		}

// 	} else {
// 		k8log.DebugLog("APIServer", "寄"+fmt.Sprint(code))
// 		k8log.DebugLog("APIServer", "寄寄寄寄寄寄寄寄寄寄寄寄寄寄寄寄寄寄寄寄寄寄寄")
// 	}

// 	return nil
// }

// // 正常运行的时候，每隔一段时间给API Server发送心跳
// func sendHeartbeat() error {
// 	k8log.InfoLog("Kublet", "try sendHeartbeat")
// 	ip, err := host.GetHostIp()
// 	if err != nil {
// 		k8log.ErrorLog("Kublet", "sendHeartbeat failed, for get host ip failed")
// 		return nil
// 	}
// 	hostName := host.GetHostName()
// 	memPercent, err := host.GetHostSystemMemoryUsage()
// 	if err != nil {
// 		k8log.ErrorLog("Kublet", "sendHeartbeat failed, for get mem percent failed")
// 		return nil
// 	}
// 	cpuPercent, err := host.GetHostSystemCPUUsage()
// 	if err != nil {
// 		k8log.ErrorLog("Kublet", "sendHeartbeat failed, for get cpu percent failed")
// 		return nil
// 	}
// 	nodeStore := apiObject.NodeStore{
// 		NodeBasic: apiObject.NodeBasic{},
// 		Status: apiObject.NodeStatus{
// 			Hostname:   hostName,
// 			Condition:  apiObject.Ready,
// 			Ip:         ip,
// 			CpuPercent: cpuPercent,
// 			MemPercent: memPercent,
// 		},
// 	}
// 	// 发送PUT请求给API Server，更新当前的Node信息
// 	perfix := kubeletconfig.DefaultKubeletConfig().APIServerURLPrefix
// 	targetAPI := perfix + config.NodesURL + hostName

// 	code, res, err := netrequest.PutRequestByTarget(targetAPI, nodeStore)
// 	if err != nil {
// 		k8log.ErrorLog("Kublet", "sendHeartbeat failed, for update node info failed "+err.Error())
// 		return nil
// 	}
// 	jsonBytes, err := json.Marshal(res)
// 	if err != nil {
// 		k8log.ErrorLog("Kublet", "sendHeartbeat failed, for marshal node info failed")
// 		return nil
// 	}
// 	k8log.InfoLog("Kublet", "sendHeartbeat resp: "+string(jsonBytes))
// 	if code == 200 {
// 		k8log.InfoLog("Kublet", "sendHeartbeat success")
// 		return nil
// 	} else {
// 		k8log.ErrorLog("Kublet", "sendHeartbeat failed, for update node info failed")
// 		return nil
// 	}
// }

// // 函数结束的时候，给API Server发送删除节点的请求
// func unRegisterNode() error {
// 	k8log.InfoLog("Kublet", "try unRegisterNode")

// 	nodeStore := apiObject.NodeStore{
// 		NodeBasic: apiObject.NodeBasic{},
// 		Status: apiObject.NodeStatus{
// 			Condition:  apiObject.Unknown,
// 			CpuPercent: -1,
// 			MemPercent: -1,
// 		},
// 	}

// 	// 发送PUT请求给API Server，更新当前的Node信息
// 	perfix := kubeletconfig.DefaultKubeletConfig().APIServerURLPrefix
// 	targetAPI := perfix + config.NodesURL + host.GetHostName()
// 	code, res, err := netrequest.PutRequestByTarget(targetAPI, nodeStore)
// 	if err != nil {
// 		k8log.ErrorLog("Kublet", "unRegisterNode failed, for update node info failed "+err.Error())
// 		return nil
// 	}
// 	jsonBytes, err := json.Marshal(res)
// 	if err != nil {
// 		k8log.ErrorLog("Kublet", "unRegisterNode failed, for marshal node info failed")
// 		return nil
// 	}
// 	k8log.InfoLog("Kublet", "unRegisterNode resp: "+string(jsonBytes))
// 	if code == 200 {
// 		k8log.InfoLog("Kublet", "unRegisterNode success")
// 		return nil
// 	} else {
// 		k8log.ErrorLog("Kublet", "unRegisterNode failed, for update node info failed")
// 		return nil
// 	}

// }

// func commonLogic() {
// 	for {
// 		// 每隔一段时间给API Server发送心跳
// 		sendHeartbeat()
// 		// 睡眠一段时间
// 		time.Sleep(2 * time.Second)
// 	}
// }

// func main() {
// 	registerNode()
// 	// 等待 goroutine 执行完毕
// 	var wg sync.WaitGroup

// 	// 启动 goroutine
// 	wg.Add(2)

// 	// // 创建一个通道来接收信号
// 	sigs := make(chan os.Signal, 1)
// 	// // 注册一个信号接收函数，将接收到的信号发送到通道
// 	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

// 	// // 启动一个 goroutine 处理信号
// 	go func() {
// 		// 阻塞等待收到信号
// 		<-sigs
// 		// 发送信号给退出前执行的函数
// 		unRegisterNode()
// 		os.Exit(0)
// 	}()

// 	// 再启动一个 goroutine 处理业务逻辑
// 	go func() {
// 		commonLogic()
// 		wg.Done()
// 	}()

// 	// 等待 goroutine 执行完毕
// 	wg.Wait()

// 	// 主线程退出
// 	os.Exit(0)

// }
