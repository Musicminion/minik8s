package allcontollers

import (
	"errors"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	minik8stypes "miniK8s/pkg/minik8sTypes"
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

func (rc *replicaController) GetAllReplicasetsFromAPIServer() ([]apiObject.ReplicaSetStore, error) {
	url := config.GetAPIServerURLPrefix() + config.GlobalReplicaSetsURL

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

func (rc *replicaController) routine() {
	pods, err := GetAllPodFromAPIServer()

	if err != nil {
		return
	}

	replicasets, err := rc.GetAllReplicasetsFromAPIServer()

	if err != nil {
		return
	}

	// 构造一个namespace/name的map，映射的value是replicasets的uuid
	replicasetsMap := make(map[string]string, 0)

	for _, rs := range replicasets {
		key := rs.Metadata.Namespace + "/" + rs.Metadata.Name
		replicasetsMap[key] = rs.Metadata.UUID
	}

	// 1. 遍历所有的replicasets
	for _, rs := range replicasets {
		meetRequirementPods := make([]apiObject.PodStore, 0)
		for _, pod := range pods {
			if CheckIfPodMeetRequirement(&pod, rs.Spec.Selector.MatchLabels) {
				meetRequirementPods = append(meetRequirementPods, pod)
			}
		}

		// 2. 根据pod的数量，调整replicasets的数量
		if len(meetRequirementPods) < rs.Spec.Replicas {
			// 需要增加replicasets的数量
			rc.AddReplicaPodsNums(&rs.Metadata, &rs.Spec.Template, rs.Spec.Replicas-len(meetRequirementPods))
		} else if len(meetRequirementPods) > rs.Spec.Replicas {
			// 需要减少replicasets的数量
			rc.ReduceReplicaPodsNums(meetRequirementPods, len(meetRequirementPods)-rs.Spec.Replicas)
		}

		// 3. 根据选择好的pod的状态，更新replicasets的状态
		// 注意，以上对replicaset的修改不会马上反映在replicaset的status里
		rc.UpdateReplicaSetStatus(meetRequirementPods, &rs)
	}

	// 2. 对于已经删除的replicasets，如果发现其对应的pod还存在，那么就删除这些pod
	for _, pod := range pods {
		if pod.Metadata.Labels[minik8stypes.Pod_ReplicaSet_UUID] != "" {
			if pod.Metadata.Labels[minik8stypes.Pod_ReplicaSet_Namespace] == "" {
				continue
			}
			if pod.Metadata.Labels[minik8stypes.Pod_ReplicaSet_Name] == "" {
				continue
			}

			key := pod.Metadata.Labels[minik8stypes.Pod_ReplicaSet_Namespace] + "/" + pod.Metadata.Labels[minik8stypes.Pod_ReplicaSet_Name]
			if _, ok := replicasetsMap[key]; !ok {
				// 说明这个pod对应的replicasets已经被删除了，那么就删除这个pod
				rc.ReduceReplicaPodsNums([]apiObject.PodStore{pod}, 1)
			}
		}
	}
}

// 增加或者减少pod的数量
func (rc *replicaController) AddReplicaPodsNums(replicaMeta *apiObject.Metadata, pod *apiObject.PodTemplate, num int) error {
	// 创建一个pod的对象
	newPod := apiObject.Pod{}
	newPod.Metadata = pod.Metadata
	newPod.Kind = apiObject.PodKind
	newPod.APIVersion = serverconfig.APIVersion
	newPod.Spec = pod.Spec
	newPod.Metadata.Labels[minik8stypes.Pod_ReplicaSet_Name] = replicaMeta.Name
	newPod.Metadata.Labels[minik8stypes.Pod_ReplicaSet_Namespace] = replicaMeta.Namespace
	newPod.Metadata.Labels[minik8stypes.Pod_ReplicaSet_UUID] = replicaMeta.UUID

	originalPodName := newPod.Metadata.Name

	originalContainerNames := make([]string, 0)

	// 遍历所有的container，修改container的name
	for _, container := range newPod.Spec.Containers {
		originalContainerNames = append(originalContainerNames, container.Name)
	}

	// 通过api server创建pod
	url := config.GetAPIServerURLPrefix() + config.PodsURL

	errStr := ""
	for i := 0; i < num; i++ {
		newPod.Metadata.Name = originalPodName + "-" + stringutil.GenerateRandomStr(5)

		// 修改container的name
		for index := range newPod.Spec.Containers {
			newPod.Spec.Containers[index].Name = originalContainerNames[index] + "-" + stringutil.GenerateRandomStr(5)
		}

		url = stringutil.Replace(url, config.URL_PARAM_NAMESPACE_PART, pod.Metadata.Namespace)
		// url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, pod.Metadata.Name+"-"+strconv.Itoa(i)+"-"+stringutil.GenerateRandomStr(5))
		code, _, err := netrequest.PostRequestByTarget(url, &newPod)

		if err != nil {
			k8log.ErrorLog("replicaController", "AddPodsNums"+err.Error())
			errStr += err.Error()
		}

		if code != http.StatusCreated {
			k8log.ErrorLog("replicaController", "AddPodsNums code is not 201")
			errStr += "code is not 200"
		}
	}

	if errStr != "" {
		return errors.New(errStr)
	}

	return nil
}

func (rc *replicaController) ReduceReplicaPodsNums(meetRequirePods []apiObject.PodStore, num int) error {
	// 遍历删除pod
	if len(meetRequirePods) < num {
		return errors.New("reduce pod nums failed")
	}

	for i := 0; i < num; i++ {
		url := config.GetAPIServerURLPrefix() + config.PodSpecURL
		url = stringutil.Replace(url, config.URL_PARAM_NAMESPACE_PART, meetRequirePods[i].Metadata.Namespace)
		url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, meetRequirePods[i].Metadata.Name)

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
			Type:           apiObject.PodKind,
			Status:         pod.Status.Phase,
			LastUpdateTime: time.Now(),
		})
	}

	newReplicaSetStatus.Replicas = replicaset.Spec.Replicas
	newReplicaSetStatus.ReadyReplicas = ReadyNums

	// 2. 更新replicaset的状态
	url := stringutil.Replace(config.ReplicaSetSpecStatusURL, config.URL_PARAM_NAMESPACE_PART, replicaset.Metadata.Namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, replicaset.Metadata.Name)
	url = config.GetAPIServerURLPrefix() + url

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
