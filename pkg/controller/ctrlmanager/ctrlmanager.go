package ctrlmanager

import "miniK8s/pkg/controller/allcontollers"

type CtrlManager interface {
	Run()
}

type ctrlManager struct {
	jobController     allcontollers.JobController
	replicaController allcontollers.ReplicaController
}

func NewCtrlManager() CtrlManager {
	newjc, err := allcontollers.NewJobController()

	if err != nil {
		panic(err)
	}

	newrc, err := allcontollers.NewReplicaController()

	if err != nil {
		panic(err)
	}

	return &ctrlManager{
		jobController:     newjc,
		replicaController: newrc,
	}
}

func (cm *ctrlManager) Run() {
	// TODO
	go cm.jobController.Run()
	cm.replicaController.Run()
}
