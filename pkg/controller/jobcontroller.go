package jobcontroller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"minik8s/cmd/kube-controller-manager/util"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/client"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/klog"
	"minik8s/pkg/listerwatcher"
	concurrentmap "minik8s/util/map"
	"path"
	"time"
)

type JobController struct {
	ls            *listerwatcher.ListerWatcher
	jobMap        *concurrentmap.ConcurrentMapTrait[string, object.VersionedGPUJob]
	jobStatusMap  *concurrentmap.ConcurrentMapTrait[string, object.VersionedJobStatus]
	apiServerBase string
	stopChannel   chan struct{}
	allocator     *object.AccountAllocator
}

func NewJobController(controllerCtx util.ControllerContext) *JobController {
	jc := &JobController{
		ls:            controllerCtx.Ls,
		stopChannel:   make(chan struct{}),
		jobMap:        concurrentmap.NewConcurrentMapTrait[string, object.VersionedGPUJob](),
		jobStatusMap:  concurrentmap.NewConcurrentMapTrait[string, object.VersionedJobStatus](),
		apiServerBase: "http://" + controllerCtx.MasterIP + ":" + controllerCtx.HttpServerPort,
		allocator:     object.NewAccountAllocator(),
	}
	if jc.apiServerBase == "" {
		klog.Fatalf("uninitialized apiserver base!\n")
	}
	return jc
}

func (jc *JobController) Run(ctx context.Context) {
	klog.Debugf("[JobController] running...\n")
	jc.register()
	<-ctx.Done()
	close(jc.stopChannel)
}

func (jc *JobController) register() {
	registerPutJob := func() {
		for {
			err := jc.ls.Watch("/registry/job/default", jc.putJob, jc.stopChannel)
			if err != nil {
				klog.Errorf("Error watching /registry/job\n")
			} else {
				return
			}
			time.Sleep(5 * time.Second)
		}
	}

	registerDelJob := func() {
		for {
			err := jc.ls.Watch("/registry/job/default", jc.delJob, jc.stopChannel)
			if err != nil {
				klog.Errorf("Error watching /registry/job\n")
			} else {
				return
			}
			time.Sleep(5 * time.Second)
		}
	}

	go registerPutJob()
	go registerDelJob()
}

func (jc *JobController) putJob(res etcdstore.WatchRes) {
	if res.ResType != etcdstore.PUT {
		return
	}
	// TODO
	job := object.GPUJob{}
	err := json.Unmarshal(res.ValueBytes, &job)
	if err != nil {
		klog.Errorf("%s\n", err.Error())
		return
	}
	account, err := jc.allocator.Allocate(job.Spec.SlurmConfig.Partition)
	if err != nil {
		klog.Errorf("%s\n", err.Error())
		return
	}
	pod := object.PodTemplate{
		ObjectMeta: object.ObjectMeta{
			Name:   fmt.Sprintf("Job-%s-Pod", job.Metadata.UID),
			Labels: map[string]string{"kind": "gpu"},
			UID:    uuid.New().String(),
		},
		Spec: object.PodSpec{
			Volumes: []object.Volume{
				{
					Name: "gpuPath",
					Type: "hostPath",
					Path: path.Join(config.SharedDataDirectory, path.Base(res.Key)),
				},
			},
			Containers: []object.Container{
				{
					Name:    "gpuPod",
					Image:   "chn1234wanghaotian/remote-runner:6.0",
					Command: nil,
					Args: []string{
						"/root/remote_runner",
						account.GetUsername(),
						account.GetPassword(),
						account.GetHost(),
						"/home/job",
						path.Join(account.GetRemoteBasePath(), path.Base(res.Key)),
					},
					VolumeMounts: []object.VolumeMount{
						{
							Name:      "gpuPath",
							MountPath: "/home/job",
						},
					},
					Ports: []object.Port{
						{ContainerPort: "9990"},
					},
				},
			},
			NodeName: "",
		},
	}
	go func() {
		time.Sleep(time.Second * 3)
		err = client.Put(jc.apiServerBase+config.PodConfigPREFIX+"/"+pod.Name, pod)
		if err != nil {
			klog.Errorf("Put job pod config error : %s\n", err.Error())
			return
		}
		err = client.Put(jc.apiServerBase+path.Join(config.Job2PodPrefix, path.Base(res.Key)), object.Job2Pod{PodName: pod.Name})
		if err != nil {
			klog.Errorf("Put Job2Pod error : %s\n", err.Error())
		}
	}()
}

func (jc *JobController) delJob(res etcdstore.WatchRes) {
	if res.ResType != etcdstore.DELETE {
		return
	}
	// TODO
}
