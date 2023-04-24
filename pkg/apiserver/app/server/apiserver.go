package apiserver

import (
	"fmt"
	"io"
	"miniK8s/pkg/apiserver/app/handlers"
	serverConfig "miniK8s/pkg/apiserver/serverconfig"
	config "miniK8s/pkg/config"
	"miniK8s/pkg/k8log"

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
}

func New(c *serverConfig.ServerConfig) ApiServer {
	gin.DefaultWriter = io.Discard
	return &apiServer{
		router:   gin.Default(),
		port:     c.Port,
		listenIP: c.ListenIP,
		ifDebug:  c.IfDebug,
	}
}

type ResponseData struct {
	Data interface{} `json:"data"`
}

// func (s *apiServer) posting(c *gin.Context) {

// }

// func (s *apiServer) putting(c *gin.Context) {}

// func (s *apiServer) deleting(c *gin.Context) {
// 	key := c.Param("key")
// 	err := s.etcdStore.Del(key)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": err.Error(),
// 		})
// 		return
// 	}
// }

// func (s *apiServer) getting(c *gin.Context) {
// key := c.Param("key")
// val, err := s.etcdStore.Get(key)
// if err != nil {
// 	c.JSON(http.StatusInternalServerError, gin.H{
// 		"error": err.Error(),
// 	})
// 	return
// }
// response := ResponseData{
// 	Data: val,
// }
// print(response.Data)
// c.JSON(http.StatusOK, response)
// }

// 不同的url, for test only
// s.router.GET("/get/:key", s.getting)
// s.router.POST("/post/:key", s.posting)
// s.router.PUT("/put/:key", s.putting)
// s.router.DELETE("/del/:key", s.deleting)

// s.router.GET("/pods",)
// s.router.GET("/", handlers.TestHandler1)
// s.router.GET(config.NodeURLWithSpecifiedName, handlers.TestHandler2)

func (s *apiServer) bind() {

	// Rest风格的api
	// 在Kubernetes API中，节点（Node）的标识符是其名称，因此在API URI中，
	// 节点的名称用于区分不同的节点。例如，获取名为node-1的节点的状态，可以使用以下URI：
	s.router.GET(config.NodesURL, handlers.GetNodes)
	s.router.GET(config.NodeSpecURL, handlers.GetNode)
	s.router.POST(config.NodesURL, handlers.AddNode)
	s.router.PUT(config.NodeSpecURL, handlers.UpdateNode)
	s.router.DELETE(config.NodeSpecURL, handlers.DeleteNode)

	// Pod相关的api
	s.router.GET(config.PodsURL, handlers.GetPods)
	s.router.POST(config.PodsURL, handlers.AddPod)

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
	s.router.Run("0.0.0.0:8090")
}
