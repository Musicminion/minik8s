package main

import "miniK8s/pkg/controller/ctrlmanager"

func main() {
	ctrlManager := ctrlmanager.NewCtrlManager()

	ctrlManager.Run()
}
