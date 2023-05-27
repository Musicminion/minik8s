package main

import (
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/serveless/server"
)

func main() {
	k8log.InfoLog("Serveless", "Serveless start")
	server := server.NewServer()

	server.Run()
}
