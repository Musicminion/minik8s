package handlers

import (
	"miniK8s/pkg/k8log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseData struct {
	Data interface{} `json:"data"`
}

func TestHandler1(c *gin.Context) {
	response := ResponseData{
		Data: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	print(response.Data)
	c.JSON(http.StatusOK, response)
	// return "test"
}

func TestHandler2(c *gin.Context) {
	name := c.Params.ByName("name")
	response := ResponseData{
		Data: name + "是你请求的参数哦！",
	}
	k8log.DebugLog("APIServer", name)
	c.JSON(http.StatusOK, response)
	// return "test"
}
