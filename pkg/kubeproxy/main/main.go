package main

import(
	"miniK8s/pkg/kubeproxy/app"
	"miniK8s/pkg/listwatcher"
)

func main() {
	proxy.NewKubeProxy(listwatcher.DefaultListwatcherConfig()).Run()
}
