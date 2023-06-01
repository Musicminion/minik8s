package main

import (
	"log"
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
)

func echoHostName(c *gin.Context) {
	// 返回 echo /etc/hostname的返回值
	out, err := exec.Command("cat", "/etc/hostname").Output()
	if err != nil {
		c.String(http.StatusOK, "echo hostname err: "+err.Error())
		return
	}
	c.String(http.StatusOK, "hostname: "+string(out))
}

func main() {
	r := gin.Default()
	r.GET("/", echoHostName)

	log.Fatal(r.Run(":8090"))
}
