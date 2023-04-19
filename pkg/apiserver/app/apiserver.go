package apiserver

import (
	// "encoding/json"
	"miniK8s/pkg/etcd"
	"miniK8s/pkg/apiserver/config"
	"net/http"
	"github.com/gin-gonic/gin"
	// "net/http"
)

type ApiServer interface {
	// Run()
}

type apiServer struct {
	router     		*gin.Engine
	etcdStore   	*etcd.Store
	port			int
}

func New(c *config.ServerConfig) ApiServer {
	store, _ := etcd.NewEtcdStore(c.EtcdEndpoints, c.EtcdTimeout)
	return &apiServer{
		router: gin.Default(),
		etcdStore: store,
		port: c.Port,
	}
}

type ResponseData struct {
	Data interface{} `json:"data"`
}

func (s *apiServer) getting(c *gin.Context) {
	key := c.Param("key")
	val, err := s.etcdStore.Get(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	response := ResponseData{
		Data: val,
	}
	print(response.Data)
	c.JSON(http.StatusOK, response)
}

func (s *apiServer) posting(c *gin.Context) {

}

func (s *apiServer) putting(c *gin.Context)  {}

func (s *apiServer) deleting(c *gin.Context) {
	key := c.Param("key")
	err := s.etcdStore.Del(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
}

func (s *apiServer) bind() {
	// 不同的url, for test only
	s.router.GET("/get/:key", s.getting)
	s.router.POST("/post/:key", s.posting)
	s.router.PUT("/put/:key", s.putting)
	s.router.DELETE("/del/:key", s.deleting)
}



func (s *apiServer) Run() {
	s.bind()
	s.router.Run(":%d", string(s.port))
}
