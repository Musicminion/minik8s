package main

import (
	"miniK8s/pkg/k8log"
	scheduler "miniK8s/pkg/scheduler/app"
)

func main() {
	scheduler, err := scheduler.NewScheduler()
	if err != nil {
		k8log.ErrorLog("scheduler", "创建调度器失败")
	}
	scheduler.Run()
}
