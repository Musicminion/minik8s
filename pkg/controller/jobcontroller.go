package jobcontroller

import (
	"context"
	"encoding/json"
	"fmt"
	"miniK8s/pkg/apiObject"
	"minik8s/cmd/kube-controller-manager/util"

	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/client"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/klog"
	"minik8s/pkg/listerwatcher"
	concurrentmap "minik8s/util/map"
	"path"
	"time"

	"github.com/google/uuid"
)

type JobController struct {
	ls            *listerwatcher.ListerWatcher
	jobMap        *concurrentmap.ConcurrentMapTrait[string, apiObjectVersionedGPUJob]
	jobStatusMap  *concurrentmap.ConcurrentMapTrait[string, apiObjectVersionedJobStatus]
	apiServerBase string
	stopChannel   chan struct{}
	allocator     *apiObjectAccountAllocator
}

func NewJobController(controllerCtx util.ControllerContext) *JobController {
	jc := &JobController{
		ls:            controllerCtx.Ls,
		stopChannel:   make(chan struct{}),
		jobMap:        concurrentmap.NewConcurrentMapTrait[string, apiObjectVersionedGPUJob](),
		jobStatusMap:  concurrentmap.NewConcurrentMapTrait[string, apiObjectVersionedJobStatus](),
		apiServerBase: "http://" + controllerCtx.MasterIP + ":" + controllerCtx.HttpServerPort,
		allocator:     apiObjectNewAccountAllocator(),
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
	job := apiObjectGPUJob{}
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
	pod := apiObjectPodTemplate{
		ObjectMeta: apiObjectObjectMeta{
			Name:   fmt.Sprintf("Job-%s-Pod", job.Metadata.UID),
			Labels: map[string]string{"kind": "gpu"},
			UID:    uuid.New().String(),
		},
		Spec: apiObjectPodSpec{
			Volumes: []apiObjectVolume{
				{
					Name: "gpuPath",
					Type: "hostPath",
					Path: path.Join(config.SharedDataDirectory, path.Base(res.Key)),
				},
			},
			Containers: []apiObjectContainer{
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
					VolumeMounts: []apiObjectVolumeMount{
						{
							Name:      "gpuPath",
							MountPath: "/home/job",
						},
					},
					Ports: []apiObjectPort{
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
		err = client.Put(jc.apiServerBase+path.Join(config.Job2PodPrefix, path.Base(res.Key)), apiObjectJob2Pod{PodName: pod.Name})
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
