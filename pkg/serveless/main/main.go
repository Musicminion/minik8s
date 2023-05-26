package main

import (
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/serveless/function"
)

func main() {
	k8log.InfoLog("Serveless", "Serveless start")
	funcController := function.NewFuncController()
	funcController.Run()
}
