package ctrlmanager

import "miniK8s/pkg/controller/allcontollers"

type CtrlManager interface {
	Run()
}

type ctrlManager struct {
	jobController allcontollers.JobController
}

func NewCtrlManager() CtrlManager {
	newjc, err := allcontollers.NewJobController()

	if err != nil {
		panic(err)
	}

	return &ctrlManager{
		jobController: newjc,
	}
}

func (cm *ctrlManager) Run() {
	// TODO
	cm.jobController.Run()
}
