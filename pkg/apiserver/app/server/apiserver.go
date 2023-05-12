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
	// "net/http"
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
	k8log.InfoLog("APIServer", "Watcher try to connect to RabbitMQ")
	go s.lw.WatchQueue_Block("apiServer", handlers.MessageHandler, make(chan struct{}))
	k8log.InfoLog("APIServer", "Bind MessageHandler To RabbitMQ Success")

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
	s.router.Run("0.0.0.0:8090")
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

	// Pod相关的api

	s.router.GET(config.PodsURL, handlers.GetPods)         // 所有pod
	s.router.GET(config.PodSpecURL, handlers.GetPod)       // 单个pod
	s.router.POST(config.PodsURL, handlers.AddPod)         // 创建pod
	s.router.PUT(config.PodSpecURL, handlers.UpdatePod)    // 更新Pod
	s.router.DELETE(config.PodSpecURL, handlers.DeletePod) // 删除Pod

	// PodStatus相关的api
	s.router.GET(config.PodSpecStatusURL, handlers.GetPodStatus)    // 获取PodStatus
	s.router.PUT(config.PodSpecStatusURL, handlers.UpdatePodStatus) // 更新PodStatus

	// Service相关的api
	s.router.POST(config.ServiceURL, handlers.AddService)       // 创建service
	s.router.GET(config.ServiceURL, handlers.GetServices)       // 获取所有service
	s.router.GET(config.ServiceSpecURL, handlers.GetService)    // 获取单个service
	s.router.PUT(config.ServiceSpecURL, handlers.UpdateService) // 更新service
}
