package test

import (
	"miniK8s/pkg/etcd"

	"github.com/gin-gonic/gin"
	// "github.com/stretchr/testify/assert"
	// "net/http"
)

//	type ApiServer interface {
//		Run()
//	}
type apiServer struct {
	router     *gin.Engine
	etcdClient *etcd.Store
}

type Result struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// func New() apiServer {
// 	return &apiServer{
// 		httpServer: gin.Default(),
// 	}
// }

// func (s *apiServer) getting(c *gin.Context) {
// 	key := "666"
// 	listRes, err := s.etcdClient.Get(key)
// 	if err != nil {
// 	}
// 	data, _ := json.Marshal(listRes[0].ValueBytes)
// 	c.JSON(http.StatusOK, gin.H{
// 		"hello": data,
// 	})
// }

// func (s *apiServer) posting(c *gin.Context) {

// }
// func (s *apiServer) putting(c *gin.Context)  {}
// func (s *apiServer) deleting(c *gin.Context) {}

// func (s *apiServer) bind() {
// 	// 不同的url
// 	s.router.GET("/someGet", s.getting)
// 	s.router.POST("/somePost", s.posting)
// 	s.router.PUT("/somePut", s.putting)
// 	s.router.DELETE("/someDelete", s.deleting)

// }

// func (api *apiServer) Run() {
// 	// 初始化etcd

// 	// 绑定HTTP request 路由？
// }

// func TestApiServer(t *testing.T){
// 	server := apiServer{}
// 	server.router = gin.Default()

// 	server.bind()
// 	err := server.router.Run(fmt.Sprintf(":%d", 8789))
// 	assert.Nil(t, err)

// 	req, _:= http.NewRequest("GET" ,"/someGet", nil)

// 	w := httptest.NewRecorder()

// 	server.router.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)
// }

// func SetupServer() apiServer {
// 	server := apiServer{}
// 	server.router = gin.Default()
// 	server.bind()
// 	return server
// }

// func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
// 	req, _ := http.NewRequest("GET", "/someGet", nil)
// 	w := httptest.NewRecorder()
// 	r.ServeHTTP(w, req)
// 	return w
// }

// func TestHelloWorld(t *testing.T) {
// 	server := SetupServer()
// 	w := performRequest(server.router, "GET", "/someGet")

// 	print(w.Body)
// 	// assert.Equal(t, http.StatusOK, w.Code)

// 	// var response map[string]string
// 	// err := json.Unmarshal([]byte(w.Body.String()), &response)

// 	// value, exists := response["Hello"]

// 	// assert.Nil(t, err)
// 	// assert.True(t, exists)
// 	// assert.Equal(t, body["Hello"], value)
// }
