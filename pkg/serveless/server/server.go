package server

import (
	"fmt"
	"miniK8s/pkg/config"
	minik8stypes "miniK8s/pkg/minik8sTypes"
	"miniK8s/pkg/serveless/function"
	"miniK8s/pkg/serveless/workflow"
	"miniK8s/util/executor"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Server interface {
	Run()
}

type server struct {
	httpServer *gin.Engine
	routeTable map[string][]string
	// routeTable 的key是 namespace/name ，value是一个数组，数组中的每个元素是一个pod的ip地址
	// 当用户的请求到来了之后，首先会根据func的namespace/name找到对应的pod的ip地址，然后再将请求转发到这个pod上
	// 如果发现这个value为空，那就需要创建一个新的pod，然后将请求转发到这个pod上

	// func的controller
	funcController function.FuncController

	// workflow的controller
	workflowController workflow.WorkflowController
}

func NewServer() Server {
	return &server{
		httpServer:         gin.Default(),
		routeTable:         make(map[string][]string),
		funcController:     function.NewFuncController(),
		workflowController: workflow.NewWorkflowController(),
	}
}

// 周期性的函数都放在这里

// 从API Server获取所有的Pod信息，根据Label筛选出所有的Function Pod
func (s *server) updateRouteTableFromAPIServer() {
	// TODO
	pods, err := GetAllPodFromAPIServer()

	if err != nil {
		return
	}

	remoteData := make(map[string][]string)

	// 遍历所有的pod，将其加入到routeTable中
	for _, pod := range pods {
		// 说明是一个Function Pod
		if pod.Metadata.Labels[minik8stypes.Pod_Func_Uuid] != "" {
			funcName := pod.Metadata.Labels[minik8stypes.Pod_Func_Name]
			funcNamespace := pod.Metadata.Labels[minik8stypes.Pod_Func_Namespace]
			key := funcNamespace + "/" + funcName
			// 检查ip是否为空
			if pod.Status.PodIP != "" {
				remoteData[key] = append(remoteData[key], pod.Status.PodIP)
				// 检查routeTable中是否有这个ip
				ifExist := false
				for _, ip := range s.routeTable[key] {
					if ip == pod.Status.PodIP {
						ifExist = true
						break
					}
				}

				if !ifExist {
					// ip不为空，说明这个pod已经启动了，可以将其加入到routeTable中
					s.routeTable[key] = append(s.routeTable[key], pod.Status.PodIP)
					fmt.Println("update routeTable: ", s.routeTable)
				}
			}
		}
	}

	// 遍历本地的routeTable，检查是否有需要删除的ip
	for key, ips := range s.routeTable {
		for _, ip := range ips {
			// 检查ip是否在remoteData中
			ifExist := false
			for _, remoteIp := range remoteData[key] {
				if remoteIp == ip {
					ifExist = true
					break
				}
			}

			if !ifExist {
				// 说明这个ip已经不存在了，需要将其删除
				fmt.Println("delete ip: ", ip)
				for i, localIp := range s.routeTable[key] {
					if localIp == ip {
						s.routeTable[key] = append(s.routeTable[key][:i], s.routeTable[key][i+1:]...)
						break
					}
				}
			}
		}
	}
}

// // 周期运行函数检查调用的function情况
// func (s *server) CheckUnusedFunc() {
// 	// TODO 遍历map callRecord，检查是否有没有被调用的function
// 	for _, record := range s.callRecord {
// 		// 检查record的EndTime是否在当前时间之前
// 		if record.EndTime.Before(time.Now()) {
// 			// 说明这个function已经超时了，需要将其删除
// 			if record.FuncCallTime == 0 {
// 				// 执行缩容的操作

// 			}

// 			record.StartTime = time.Now()
// 			record.EndTime = time.Now().Add(time.Duration(5) * time.Minute)
// 			record.FuncCallTime = 0
// 		}
// 	}

// }

func (s *server) Run() {
	// 周期性的更新routeTable
	go executor.Period(RouterUpdate_Delay, RouterUpdate_WaitTime, s.updateRouteTableFromAPIServer, RouterUpdate_ifLoop)

	// 周期性的检查function的情况，如果有新创建的function，那么就创建一个新的pod
	go s.funcController.Run()

	// 周期性的检查workflow的情况，如果有新创建的workflow，那么就创建一个新的pod
	go s.workflowController.Run()

	// 初始化服务器
	s.httpServer.POST("/:funcNamespace/:funcName", s.handleFuncRequest)
	s.httpServer.GET("/:funcNamespace/:funcName", s.checkFunction)
	s.httpServer.Run(":" + strconv.Itoa(config.Serveless_Server_Port))
}
