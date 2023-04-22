package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseData struct {
	Data interface{} `json:"data"`
}

func TestHandler(c *gin.Context) {
	response := ResponseData{
		Data: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	print(response.Data)
	c.JSON(http.StatusOK, response)
	// return "test"
}
