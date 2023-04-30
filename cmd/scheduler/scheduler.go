package main

import (
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/scheduler"
)

func main() {
	scheduler, err := scheduler.NewScheduler()
	if err != nil {
		k8log.ErrorLog("scheduler", "创建调度器失败")
	}
	scheduler.Run()
}
