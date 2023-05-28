package allcontollers

import (
	"errors"
	"fmt"
	"math"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	minik8stypes "miniK8s/pkg/minik8sTypes"
	"miniK8s/util/executor"
	netrequest "miniK8s/util/netRequest"
	"miniK8s/util/stringutil"
	"net/http"
	"strconv"
	"time"
)

var (
	HpaControllerUpdateDelay     = time.Second * 0
	HpaControllerUpdateFrequency = []time.Duration{15 * time.Second}
	HpaControllerUpdateLoop      = true
)

type HpaController interface {
	Run()
}

type hpaController struct {
}

func NewHpaController() (HpaController, error) {
	return &hpaController{}, nil
}

func (hc *hpaController) GetAllHpaFromAPIServer() ([]apiObject.HPAStore, error) {
	url := config.GetAPIServerURLPrefix() + config.GlobalHPAURL

	allHPA := make([]apiObject.HPAStore, 0)

	code, err := netrequest.GetRequestByTarget(url, &allHPA, "data")

	if err != nil {
		return nil, err
	}

	if code != http.StatusOK {
		return nil, errors.New("get all hpa from apiserver failed")
	}

	return allHPA, nil
}

// 将更新过的hpa存入到etcd中
func (hc *hpaController) UpdateHpaStatus(hpa apiObject.HPAStore) error {
	url := config.GetAPIServerURLPrefix() + config.HPASpecStatusURL
	url = stringutil.Replace(url, config.URL_PARAM_NAMESPACE_PART, hpa.Metadata.Namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, hpa.Metadata.Name)

	code, _, err := netrequest.PutRequestByTarget(url, &(hpa.Status))

	if err != nil {
		return err
	}
	if code != http.StatusOK {
		return errors.New("update hpa status failed, code expected 200, but got " + strconv.Itoa(code))
	}

	return nil
}

func (hc *hpaController) AddOneHpaPod(hpa apiObject.HPAStore, podTemplate apiObject.Pod) error {
	k8log.DebugLog("HpaController", fmt.Sprintf("AddOneHpaPod: hpa=%s, pod=%s", hpa.Metadata.Name, podTemplate.Metadata.Name))
	// 根据podTemplate，创建新的pod
	newPod := podTemplate
	newPod.Metadata.Name = podTemplate.GetObjectName() + "-" + stringutil.GenerateRandomStr(5)

	// 修改container的name
	for index := range podTemplate.Spec.Containers {
		newPod.Spec.Containers[index].Name = podTemplate.Spec.Containers[index].Name + "-" + stringutil.GenerateRandomStr(5)
	}

	// 为新的label打上hpa的标签
	newPod.Metadata.Labels[minik8stypes.Pod_HPA_Name] = hpa.Metadata.Name
	newPod.Metadata.Labels[minik8stypes.Pod_HPA_Namespace] = hpa.Metadata.Namespace
	newPod.Metadata.Labels[minik8stypes.Pod_HPA_UUID] = hpa.Metadata.UUID

	// 通过api server创建pod
	url := config.GetAPIServerURLPrefix() + config.PodsURL
	url = stringutil.Replace(url, config.URL_PARAM_NAMESPACE_PART, hpa.Spec.Workload.Metadata.Namespace)
	code, _, err := netrequest.PostRequestByTarget(url, &newPod)
	if err != nil {
		k8log.ErrorLog("hpaController", "AddHpaPods "+err.Error())
		return err
	}
	if code != http.StatusCreated {
		k8log.ErrorLog("hpaController", "AddHpaPods code not 201")
		return err
	}
	return nil
}

func (hc *hpaController) ReduceOneHpaPod(pod apiObject.PodStore) error {
	k8log.DebugLog("HpaController", fmt.Sprintf("ReduceOneHpaPod: pod=%s", pod.Metadata.Name))
	// 通过api server删除pod
	url := config.GetAPIServerURLPrefix() + config.PodSpecURL
	url = stringutil.Replace(url, config.URL_PARAM_NAMESPACE_PART, pod.Metadata.Namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, pod.Metadata.Name)

	code, err := netrequest.DelRequest(url)
	if err != nil {
		k8log.ErrorLog("hpaController", "ReducePods "+err.Error())
		return err
	}
	if code != http.StatusNoContent {
		k8log.ErrorLog("hpaController", "ReducePods code not 204")
		return err
	}
	return nil
}

func (hc *hpaController) HandleHPAUpdate(hpa apiObject.HPAStore, pods []apiObject.PodStore) {
	// 1. 根据hpa的selector，找到所有满足要求的pod
	meetRequirementPods := make([]apiObject.PodStore, 0)
	for _, pod := range pods {
		if CheckIfPodMeetRequirement(&pod, hpa.Spec.Selector.MatchLabels) {
			meetRequirementPods = append(meetRequirementPods, pod)
		}
	}
	hpa.Status.CurrentReplicas = len(meetRequirementPods)

	// 2. 判断hpa的pod是否在规定区间之内, 如果不在，需要扩容或者缩容
	if hpa.Status.CurrentReplicas < hpa.Spec.MinReplicas {
		err := hc.AddOneHpaPod(hpa, *meetRequirementPods[0].ToPod())
		if err != nil {
			k8log.ErrorLog("hpaController", "HandleHPAUpdate "+err.Error())
		}
		return
	}
	if hpa.Status.CurrentReplicas > hpa.Spec.MaxReplicas {
		err := hc.ReduceOneHpaPod(meetRequirementPods[0])
		if err != nil {
			k8log.ErrorLog("hpaController", "HandleHPAUpdate "+err.Error())
		}
		return
	}

	// 3. 计算hpa匹配的pod的平均cpu和memory使用率
	averageCPUUsage := hc.CalculateAverageCPUUsage(meetRequirementPods)
	averageMemoryUsage := hc.CalculateAverageMemoryUsage(meetRequirementPods)

	// 4. 根据hpa的spec和计算出来的平均使用率，得到期望的replica个数
	expectedReplicas := hc.CalculateExpectedReplicas(hpa, averageCPUUsage, averageMemoryUsage)
	if expectedReplicas > hpa.Status.CurrentReplicas {
		hc.AddOneHpaPod(hpa, *meetRequirementPods[0].ToPod())
	}
	if expectedReplicas < hpa.Status.CurrentReplicas {
		hc.ReduceOneHpaPod(meetRequirementPods[0])
	}

	// 5. 更新hpa的status, replica的数量不会马上更新，因为需要时间创建或者删除pod
	// 这会在下一次update时得到更新（如果操作已经完成）
	hpa.Status.CurCPUPercent = averageCPUUsage
	hpa.Status.CurMemPercent = averageMemoryUsage
	err := hc.UpdateHpaStatus(hpa)
	if err != nil {
		k8log.ErrorLog("hpaController", "HandleHPAUpdate "+err.Error())
	}

}

func (hc *hpaController) Routine() {
	pods, err := GetAllPodFromAPIServer()
	if err != nil {
		return
	}
	if len(pods) == 0 {
		k8log.DebugLog("hpaController", "hpa can't find any pods")
		return
	}

	hpas, err := hc.GetAllHpaFromAPIServer()
	if err != nil {
		return
	}

	// 构造一个namespace/name的map，映射的value是hpa的uuid
	hpaMap := make(map[string]string, 0)

	for _, hpa := range hpas {
		key := hpa.Metadata.Namespace + "/" + hpa.Metadata.Name
		hpaMap[key] = hpa.Metadata.UUID
	}

	// 遍历所有的hpa，执行update操作
	for _, hpa := range hpas {
		go hc.HandleHPAUpdate(hpa, pods)
	}

	// 对于已经删除的hpa，如果发现其对应的pod还存在，那么就删除这些pod
	for _, pod := range pods {
		if pod.Metadata.Labels[minik8stypes.Pod_HPA_UUID] != "" {
			if pod.Metadata.Labels[minik8stypes.Pod_HPA_Namespace] == "" {
				continue
			}
			if pod.Metadata.Labels[minik8stypes.Pod_HPA_Name] == "" {
				continue
			}

			key := pod.Metadata.Labels[minik8stypes.Pod_HPA_Namespace] + "/" + pod.Metadata.Labels[minik8stypes.Pod_HPA_Name]
			if _, ok := hpaMap[key]; !ok {
				// 说明这个pod对应的replicasets已经被删除了，那么就删除这个pod
				hc.ReduceOneHpaPod(pod)
			}
		}
	}
}

func (hc *hpaController) CalculateAverageCPUUsage(pods []apiObject.PodStore) float64 {
	totolCPUUsage := 0.0
	for _, pod := range pods {
		totolCPUUsage += pod.Status.CpuPercent
	}
	return totolCPUUsage / float64(len(pods))
}

func (hc *hpaController) CalculateAverageMemoryUsage(pods []apiObject.PodStore) float64 {
	totolMemoryUsage := 0.0
	for _, pod := range pods {
		totolMemoryUsage += pod.Status.MemPercent
	}
	return totolMemoryUsage / float64(len(pods))
}

func (hc *hpaController) CalculateExpectedReplicas(hpa apiObject.HPAStore, cpuUsage float64, memoryUsage float64) int {
	// 根据cpu和memory的使用率，计算出来的期望的replica个数
	// 期望的replica个数是基于cpu和memory使用率算出的最大值
	cpuUsedPercent := cpuUsage / float64(hpa.Spec.Metrics.CPUPercent)
	memoryUsedPercent := memoryUsage / float64(hpa.Spec.Metrics.MemPercent)
	expectedReplicas := int(math.Max(cpuUsedPercent, memoryUsedPercent) * float64(hpa.Status.CurrentReplicas))
	k8log.DebugLog("hpaController", "memoryUsedPercent: "+strconv.FormatFloat(memoryUsedPercent, 'f', 2, 64))
	k8log.DebugLog("hpaController", "cpuUsedPercent: "+strconv.FormatFloat(cpuUsedPercent, 'f', 2, 64))
	k8log.DebugLog("hpaController", "expectedReplicas: "+strconv.Itoa(expectedReplicas))

	// 期望的replica不能越界
	if expectedReplicas > hpa.Spec.MaxReplicas {
		expectedReplicas = hpa.Spec.MaxReplicas
	}
	if expectedReplicas < hpa.Spec.MinReplicas {
		expectedReplicas = hpa.Spec.MinReplicas
	}

	return expectedReplicas
}

func (hc *hpaController) Run() {
	// 定期执行
	executor.Period(HpaControllerUpdateDelay, HpaControllerUpdateFrequency, hc.Routine, HpaControllerUpdateLoop)
}
