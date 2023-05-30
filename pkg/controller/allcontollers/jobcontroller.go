package allcontollers

import (
	"encoding/json"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/listwatcher"
	"miniK8s/pkg/message"
	netrequest "miniK8s/util/netRequest"
	"miniK8s/util/stringutil"
	"net/http"

	"github.com/streadway/amqp"
)

const (
	GPU_Server_Image = "musicminion/minik8s-gpu:latest"
)

type JobController interface {
	Run()
}

type jobController struct {
	lw *listwatcher.Listwatcher
}

func NewJobController() (JobController, error) {
	lwConfig := listwatcher.DefaultListwatcherConfig()
	newlw, err := listwatcher.NewListWatcher(lwConfig)

	if err != nil {
		return nil, err
	}

	return &jobController{
		lw: newlw,
	}, nil
}

func (jc *jobController) JobCreateHandler(parsedMsg *message.Message) {
	jobMeta := &apiObject.Basic{}
	err := json.Unmarshal([]byte(parsedMsg.Content), jobMeta)

	if err != nil {
		k8log.ErrorLog("Job-Controller", "HandleServiceUpdate: failed to unmarshal")
		return
	}

	// 主动请求job的信息
	targetURI := parsedMsg.ResourceURI
	targetURI = config.GetAPIServerURLPrefix() + targetURI

	job := &apiObject.JobStore{}
	code, err := netrequest.GetRequestByTarget(targetURI, job, "data")

	if err != nil {
		k8log.ErrorLog("Job-Controller", "HandleJobCreate: failed to get job"+err.Error())
		return
	}

	if code != http.StatusOK {
		k8log.ErrorLog("Job-Controller", "HandleJobCreate: failed to get job [Not 200]")
		return
	}

	// 不需要主动检查jobFile
	// 因为是jobfile主动发送的消息，所以jobfile一定是存在的

	// /bin/job-server -jobName YourJobName -jobNamespace YourJobNamespace -apiServerAddr YourAPIServerAddr
	// "/bin/job-server --jobName=" + job.Metadata.Name + " --jobNamespace=" + job.Metadata.Namespace + " --apiServerAddr=http://192.168.126.130:8090"
	containerCmd := []string{
		"/bin/job-server",
		"--jobName=" + job.Metadata.Name,
		"--jobNamespace=" + job.Metadata.Namespace,
		"--apiServerAddr=" + config.GetAPIServerURLPrefix(),
	}

	// 创建一个pod
	pod := &apiObject.Pod{
		Basic: apiObject.Basic{
			APIVersion: serverconfig.APIVersion,
			Kind:       apiObject.PodKind,
			Metadata: apiObject.Metadata{
				Name:      job.Metadata.Name,
				Namespace: job.Metadata.Namespace,
			},
		},
		Spec: apiObject.PodSpec{
			// NodeName: "ubuntu",   // TODO: change node
			Containers: []apiObject.Container{
				{
					Name:    "gpu-server" + job.Metadata.UUID,
					Image:   GPU_Server_Image,
					Command: containerCmd,
				},
			},
		},
	}

	podURI := stringutil.Replace(config.PodsURL, config.URL_PARAM_NAMESPACE_PART, job.Metadata.Namespace)
	podURI = config.GetAPIServerURLPrefix() + podURI

	code, _, err = netrequest.PostRequestByTarget(podURI, pod)

	if err != nil {
		k8log.ErrorLog("Job-Controller", "HandleServiceUpdate: failed to create pod"+err.Error())
		return
	}

	if code != http.StatusCreated {
		k8log.ErrorLog("Job-Controller", "HandleServiceUpdate: failed to create pod [Not 201]")
		return
	}

	k8log.InfoLog("Job-Controller", "HandleServiceUpdate: success to create pod")
}

func (jc *jobController) JobDeleteHandler(msg *message.Message) {

}

func (jc *jobController) MsgHandler(msg amqp.Delivery) {
	k8log.WarnLog("Job-Controller", "收到消息"+string(msg.Body))

	parsedMsg, err := message.ParseJsonMessageFromBytes(msg.Body)
	if err != nil {
		k8log.ErrorLog("Job-Controller", "消息格式错误,无法转换为Message")
	}

	switch parsedMsg.Type {
	// 同样的也是创建一个Pod
	case message.UPDATE:
		jc.JobCreateHandler(parsedMsg)
	// 这里的删除对应的是删除pod
	case message.DELETE:
		jc.JobDeleteHandler(parsedMsg)
	}
}

func (jc *jobController) Run() {
	jc.lw.WatchQueue_Block(message.JobUpdateQueue, jc.MsgHandler, make(chan struct{}))
}
