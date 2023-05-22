package ctrlmanager

import (
	"miniK8s/pkg/controller/allcontollers"
	"miniK8s/pkg/k8log"
)

type CtrlManager interface {
	Run(stopCh <-chan struct{})
}

type ctrlManager struct {
	jobController allcontollers.JobController
	dnsController allcontollers.DnsController
}

func NewCtrlManager() CtrlManager {
	newjc, err := allcontollers.NewJobController()
	if err != nil {
		panic(err)
	}
	newdc, err := allcontollers.NewDnsController()
	if err != nil {
		panic(err)
	}
	return &ctrlManager{
		jobController: newjc,
		dnsController: newdc,
	}
}

func (cm *ctrlManager) Run(stopCh <-chan struct{}) {
	// TODO
	go cm.jobController.Run()
	go cm.dnsController.Run()

	// wait for stop signal
	_, ok := <-stopCh
	if !ok {
		k8log.ErrorLog("CtrlManager", "stopCh closed")
	}
	k8log.InfoLog("CtrlManager", "stop signal received")


}
