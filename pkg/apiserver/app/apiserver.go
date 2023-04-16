package apiserver

import (
	"github.com/gin-gonic/gin"
	// "net/http"
)

type ApiServer interface {
	Run()
}

func New() ApiServer {
	return &apiServer{
		httpServer: gin.Default(),
	}
}

type apiServer struct {
	httpServer *gin.Engine
}

func (api *apiServer) Run() {

}
