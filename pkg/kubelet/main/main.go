package main

import (
	"miniK8s/pkg/k8log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// 初始化的时候给API Server注册当前的节点
func registerNode() error {
	k8log.InfoLog("[Kublet] registerNode")
	return nil
}

// 正常运行的时候，每隔一段时间给API Server发送心跳
func sendHeartbeat() error {
	k8log.InfoLog("[Kublet] sendHeartbeat")
	return nil
}

// 函数结束的时候，给API Server发送删除节点的请求
func unRegisterNode() error {
	k8log.InfoLog("[Kublet] unRegisterNode")
	return nil
}

func commonLogic() {
	for {
		// 每隔一段时间给API Server发送心跳
		sendHeartbeat()
		// 睡眠一段时间
		time.Sleep(5 * time.Second)
	}
}

func main() {
	// 等待 goroutine 执行完毕
	var wg sync.WaitGroup

	// 启动 goroutine
	wg.Add(1)

	// // 创建一个通道来接收信号
	sigs := make(chan os.Signal, 1)
	// // 注册一个信号接收函数，将接收到的信号发送到通道
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// // 启动一个 goroutine 处理信号
	go func() {
		// 在这里执行业务逻辑
		// ...
		registerNode()
		// 处理完业务逻辑后，阻塞等待收到信号
		<-sigs
		// 发送信号给退出前执行的函数
		unRegisterNode()
		os.Exit(0)
	}()

	// 再启动一个 goroutine 处理业务逻辑
	go func() {
		commonLogic()
		wg.Done()
	}()

	// 等待 goroutine 执行完毕
	wg.Wait()

	// 主线程退出
	os.Exit(0)

}
