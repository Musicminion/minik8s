package allcontollers

import (
	"errors"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	"miniK8s/util/executor"
	netrequest "miniK8s/util/netRequest"
	"miniK8s/util/stringutil"
	"net/http"
	"time"
)

type ReplicaController interface {
	Run()
}

type replicaController struct {
}

func NewReplicaController() (ReplicaController, error) {
	return &replicaController{}, nil
}

func (rc *replicaController) GetAllPodFromAPIServer() ([]apiObject.PodStore, error) {
	url := config.API_Server_URL_Prefix + config.GlobalPodsURL

	allPods := make([]apiObject.PodStore, 0)

	code, err := netrequest.GetRequestByTarget(url, &allPods, "data")

	if err != nil {
		return nil, err
	}

	if code != http.StatusOK {
		return nil, errors.New("get all pods from apiserver failed")
	}

	return allPods, nil
}

func (rc *replicaController) GetAllReplicasetsFromAPIServer() ([]apiObject.ReplicaSetStore, error) {
	url := config.API_Server_URL_Prefix + config.GlobalReplicaSetsURL

	allReplicaSets := make([]apiObject.ReplicaSetStore, 0)

	code, err := netrequest.GetRequestByTarget(url, &allReplicaSets, "data")

	if err != nil {
		return nil, err
	}

	if code != http.StatusOK {
		return nil, errors.New("get all replicasets from apiserver failed")
	}

	return allReplicaSets, nil
}

func (rc *replicaController) CheckIfPodMeetRequirement(pod *apiObject.PodStore, selector *apiObject.ReplicaSetSelector) bool {
	// 这里的匹配策略是：只要pod的label中有一个key-value对与selector中的key-value对相同，就认为pod满足要求
	podLabel := pod.Metadata.Labels
	for key, value := range selector.MatchLabels {
		if podLabel[key] == value {
			return true
		} else {
			continue
		}
	}

	return false
}

func (rc *replicaController) routine() {
	pods, err := rc.GetAllPodFromAPIServer()

	if err != nil {
		return
	}

	replicasets, err := rc.GetAllReplicasetsFromAPIServer()

	if err != nil {
		return
	}

	// 1. 遍历所有的replicasets
	for _, rs := range replicasets {
		meetRequirementPods := make([]apiObject.PodStore, 0)
		for _, pod := range pods {
			if rc.CheckIfPodMeetRequirement(&pod, &rs.Spec.Selector) {
				meetRequirementPods = append(meetRequirementPods, pod)
			}
		}

		// 2. 根据pod的数量，调整replicasets的数量
		if len(meetRequirementPods) < rs.Spec.Replicas {
			// 需要增加replicasets的数量
			rc.AddPodsNums(&rs.Spec.Template, rs.Spec.Replicas-len(meetRequirementPods))
		} else if len(meetRequirementPods) > rs.Spec.Replicas {
			// 需要减少replicasets的数量
			rc.ReducePodsNums(meetRequirementPods, len(meetRequirementPods)-rs.Spec.Replicas)
		}

		// 3. 根据选择好的pod的状态，更新replicasets的状态
		rc.UpdateReplicaSetStatus(meetRequirementPods, &rs)

	}
}

// 增加或者减少pod的数量
func (rc *replicaController) AddPodsNums(pod *apiObject.PodTemplate, num int) error {
	// 创建一个pod的对象
	newPod := apiObject.Pod{}
	newPod.Metadata = pod.Metadata
	newPod.Spec = pod.Spec

	// 通过api server创建pod
	url := config.API_Server_URL_Prefix + config.PodsURL

	errStr := ""
	for i := 0; i < num; i++ {
		code, _, err := netrequest.PostRequestByTarget(url, &newPod)

		if err != nil {
			k8log.ErrorLog("replicaController", "AddPodsNums"+err.Error())
			errStr += err.Error()
		}

		if code != http.StatusCreated {
			k8log.ErrorLog("replicaController", "AddPodsNums code is not 200")
			errStr += "code is not 200"
		}
	}

	if errStr != "" {
		return errors.New(errStr)
	}

	return nil
}

func (rc *replicaController) ReducePodsNums(meetRequirePods []apiObject.PodStore, num int) error {
	// 遍历删除pod
	if len(meetRequirePods) < num {
		return errors.New("reduce pod nums failed")
	}

	for i := 0; i < num; i++ {
		url := config.PodSpecURL
		url = stringutil.Replace(url, config.URL_PARAM_NAMESPACE_PART, meetRequirePods[i].Metadata.Namespace)
		url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, meetRequirePods[i].Metadata.Name)
		url = config.API_Server_URL_Prefix + url

		code, err := netrequest.DelRequest(url)

		if err != nil {
			k8log.ErrorLog("replicaController", "ReducePodsNums"+err.Error())
		}

		if code != http.StatusNoContent {
			k8log.ErrorLog("replicaController", "ReducePodsNums code is not 204")
		}
	}

	return nil
}

func (rc *replicaController) UpdateReplicaSetStatus(meetRequirePods []apiObject.PodStore, replicaset *apiObject.ReplicaSetStore) error {
	// 遍历所有的pod，获取所有pod的状态
	// 创建一个replicasetStatus对象
	newReplicaSetStatus := apiObject.ReplicaSetStatus{}
	// 创建一个replicasetStatus.Conditions
	newReplicaSetStatus.Conditions = make([]apiObject.ReplicaSetCondition, 0)

	ReadyNums := 0
	// 1. 遍历所有的pod，获取所有pod的状态
	for _, pod := range meetRequirePods {

		if pod.Status.Phase == apiObject.PodRunning {
			ReadyNums += 1
		}

		newReplicaSetStatus.Conditions = append(newReplicaSetStatus.Conditions, apiObject.ReplicaSetCondition{
			Type:           "Pod",
			Status:         pod.Status.Phase,
			LastUpdateTime: time.Now(),
		})
	}

	newReplicaSetStatus.Replicas = replicaset.Spec.Replicas
	newReplicaSetStatus.ReadyReplicas = ReadyNums

	// 2. 更新replicaset的状态
	url := stringutil.Replace(config.ReplicaSetSpecStatusURL, config.URL_PARAM_NAMESPACE_PART, replicaset.Metadata.Namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, replicaset.Metadata.Name)
	url = config.API_Server_URL_Prefix + url

	code, _, err := netrequest.PutRequestByTarget(url, &newReplicaSetStatus)

	if err != nil {
		k8log.ErrorLog("replicaController", "UpdateReplicaSetStatus"+err.Error())
		return err
	}

	if code != http.StatusOK {
		k8log.ErrorLog("replicaController", "UpdateReplicaSetStatus code is not 200")
		return errors.New("UpdateReplicaSetStatus code is not 200")
	}

	return nil
}

func (rc *replicaController) Run() {
	// 定期执行
	executor.Period(ReplicaControllerUpdateDelay, ReplicaControllerUpdateFrequency, rc.routine, ReplicaControllerUpdateLoop)
}

var (
	ReplicaControllerUpdateDelay     = 5 * time.Second
	ReplicaControllerUpdateFrequency = []time.Duration{10 * time.Second}
	ReplicaControllerUpdateLoop      = true
)
