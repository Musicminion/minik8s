package main

import (
	scheduler "miniK8s/pkg/scheduler/app"
)

// 启动调度器
func main() {
	// 创建一个调度器
	sch, err := scheduler.NewScheduler()
	if err != nil {
		panic(err)
	}

	sch.Run()
}
