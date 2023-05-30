package apiserver

import (
	"fmt"
	"io"
	"miniK8s/pkg/apiserver/app/handlers"
	serverConfig "miniK8s/pkg/apiserver/serverconfig"
	config "miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/listwatcher"
	"github.com/gin-gonic/gin"
)

type ApiServer interface {
	Run()
}

type apiServer struct {
	router   *gin.Engine
	listenIP string
	port     int
	ifDebug  bool
	lw       *listwatcher.Listwatcher
}

func New(c *serverConfig.ServerConfig) ApiServer {
	gin.DefaultWriter = io.Discard
	lw, err := listwatcher.NewListWatcher(listwatcher.DefaultListwatcherConfig())
	if err != nil {
		k8log.FatalLog("apiserver", fmt.Sprintf("创建ListWatcher失败:%s", err.Error()))
	}

	return &apiServer{
		router:   gin.Default(),
		port:     c.Port,
		listenIP: c.ListenIP,
		ifDebug:  c.IfDebug,
		lw:       lw,
	}
}

type ResponseData struct {
	Data interface{} `json:"data"`
}

func (s *apiServer) Run() {
	k8log.InfoLog("APIServer", "Starting api server")
	if s.ifDebug {
		gin.SetMode(gin.DebugMode)
		k8log.InfoLog("APIServer", "Debug mode is on")
	} else {
		gin.SetMode(gin.ReleaseMode)
		k8log.InfoLog("APIServer", "Debug mode is off, release mode is on")
	}

	s.bind()
	runAddr := s.listenIP + ":" + fmt.Sprint(s.port)
	k8log.InfoLog("APIServer", "Listening on "+runAddr)
	s.router.Run("0.0.0.0:" + fmt.Sprint(s.port))
}

func (s *apiServer) bind() {

	// Rest风格的api
	// 在Kubernetes API中，节点（Node）的标识符是其名称，因此在API URI中，
	// 节点的名称用于区分不同的节点。例如，获取名为node-1的节点的状态，可以使用以下URI：
	s.router.GET(config.NodesURL, handlers.GetNodes)
	s.router.GET(config.NodeSpecURL, handlers.GetNode)
	s.router.POST(config.NodesURL, handlers.AddNode)
	s.router.PUT(config.NodeSpecURL, handlers.UpdateNode)
	s.router.DELETE(config.NodeSpecURL, handlers.DeleteNode)

	// 对于节点的状态
	s.router.GET(config.NodeSpecStatusURL, handlers.GetNodeStatus)
	s.router.PUT(config.NodeSpecStatusURL, handlers.UpdateNodeStatus)

	// 节点的Pod
	s.router.GET(config.NodeAllPodsURL, handlers.GetNodePods)

	// Pod相关的api
	s.router.GET(config.GlobalPodsURL, handlers.GetGlobalPods) // 所有pod
	s.router.GET(config.PodsURL, handlers.GetPods)             // 所有pod
	s.router.GET(config.PodSpecURL, handlers.GetPod)           // 单个pod
	s.router.POST(config.PodsURL, handlers.AddPod)             // 创建pod
	s.router.PUT(config.PodSpecURL, handlers.UpdatePod)        // 更新Pod
	s.router.DELETE(config.PodSpecURL, handlers.DeletePod)     // 删除Pod

	// PodStatus相关的api
	s.router.GET(config.PodSpecStatusURL, handlers.GetPodStatus)     // 获取PodStatus
	s.router.POST(config.PodSpecStatusURL, handlers.UpdatePodStatus) // 更新PodStatus

	// Service相关的api
	s.router.POST(config.ServiceURL, handlers.AddService)          // 创建service
	s.router.GET(config.ServiceURL, handlers.GetServices)          // 获取所有service
	s.router.GET(config.ServiceSpecURL, handlers.GetService)       // 获取单个service
	s.router.PUT(config.ServiceSpecURL, handlers.UpdateService)    // 更新service
	s.router.DELETE(config.ServiceSpecURL, handlers.DeleteService) // 删除service

	// Job相关的api
	s.router.GET(config.JobsURL, handlers.GetJobs)         // 获取所有job
	s.router.GET(config.JobSpecURL, handlers.GetJob)       // 获取单个job
	s.router.POST(config.JobsURL, handlers.AddJob)         // 创建job
	s.router.DELETE(config.JobSpecURL, handlers.DeleteJob) // 删除job

	// JobStatus相关的api
	s.router.GET(config.JobSpecStatusURL, handlers.GetJobStatus)    // 获取jobStatus
	s.router.PUT(config.JobSpecStatusURL, handlers.UpdateJobStatus) // 更新jobStatus

	// JobFile相关的api
	s.router.GET(config.JobFileSpecURL, handlers.GetJobFile) // 获取jobFile
	s.router.POST(config.JobFileURL, handlers.AddJobFile)    // 创建jobFile

	s.router.PUT(config.JobFileSpecURL, handlers.UpdateJobFile) // 更新jobFile

	// replicaSet相关的api
	s.router.GET(config.GlobalReplicaSetsURL, handlers.GetReplicaSets)   // 获取所有replicaSet
	s.router.GET(config.ReplicaSetsURL, handlers.GetReplicaSets)         // 获取名字空间下面的所有replicaSet
	s.router.GET(config.ReplicaSetSpecURL, handlers.GetReplicaSet)       // 获取单个replicaSet
	s.router.POST(config.ReplicaSetsURL, handlers.AddReplicaSet)         // 创建replicaSet
	s.router.DELETE(config.ReplicaSetSpecURL, handlers.DeleteReplicaSet) // 删除replicaSet
	s.router.PUT(config.ReplicaSetSpecURL, handlers.UpdateReplicaSet)    // 更新replicaSet

	//
	s.router.GET(config.ReplicaSetSpecStatusURL, handlers.GetReplicaSetStatus)    // 获取replicaSetStatus
	s.router.PUT(config.ReplicaSetSpecStatusURL, handlers.UpdateReplicaSetStatus) // 更新replicaSetStatus

	// Dns相关的api
	s.router.GET(config.DnsURL, handlers.GetDnsList)       // 获取所有dns
	s.router.GET(config.DnsSpecURL, handlers.GetDns)       // 获取单个dns
	s.router.POST(config.DnsURL, handlers.AddDns)          // 创建dns
	s.router.DELETE(config.DnsSpecURL, handlers.DeleteDns) // 删除dns

	// HPA相关的api
	s.router.GET(config.HPAURL, handlers.GetHPAs)             // 获取所有hpa
	s.router.GET(config.HPASpecURL, handlers.GetHPA)          // 获取单个hpa
	s.router.POST(config.HPAURL, handlers.AddHPA)             // 创建hpa
	s.router.PUT(config.HPASpecURL, handlers.UpdateHPA)       // 更新hpa
	s.router.DELETE(config.HPASpecURL, handlers.DeleteHPA)    // 删除hpa
	s.router.GET(config.GlobalHPAURL, handlers.GetGlobalHPAs) // 获取所有hpa

	// Function相关的api
	s.router.GET(config.FunctionURL, handlers.GetFunctions)       // 获取所有function
	s.router.GET(config.FunctionSpecURL, handlers.GetFunction)    // 获取单个function
	s.router.POST(config.FunctionURL, handlers.AddFunction)       // 创建function
	s.router.PUT(config.FunctionSpecURL, handlers.UpdateFunction) // 更新function
	s.router.DELETE(config.FunctionSpecURL, handlers.DeleteFunction)
	s.router.GET(config.GlobalFunctionsURL, handlers.GetGlobalFunctions) // 获取所有function

	s.router.PUT(config.HPASpecStatusURL, handlers.UpdateHPAStatus) // 更新hpaStatus

	// WorkFlow相关的api
	s.router.GET(config.GlobalWorkflowsURL, handlers.GetGlobalWorkFlows) // 获取所有WorkFlow
	s.router.GET(config.WorkflowURL, handlers.GetWorkFlows)              // 获取所有WorkFlow
	s.router.GET(config.WorkflowSpecURL, handlers.GetWorkFlow)           // 获取单个WorkFlow
	s.router.POST(config.WorkflowURL, handlers.AddWorkFlow)              // 创建WorkFlow
	s.router.PUT(config.WorkflowSpecURL, handlers.UpdateWorkFlow)        // 更新WorkFlow
	s.router.DELETE(config.WorkflowSpecURL, handlers.DeleteWorkFlow)     // 删除WorkFlow

	// WorkFlowStatus相关的api
	s.router.GET(config.WorkflowSpecStatusURL, handlers.GetWorkFlowStatus)    // 获取WorkFlowStatus
	s.router.PUT(config.WorkflowSpecStatusURL, handlers.UpdateWorkFlowStatus) // 更新WorkFlowStatus

}
