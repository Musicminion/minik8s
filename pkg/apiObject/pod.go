package apiObject

import (
	"time"

	"github.com/docker/docker/api/types"
)

type ContainerPort struct {
	Name          string `yaml:"name" json:"name"`
	HostPort      string `yaml:"hostPort" json:"hostPort"`
	ContainerPort string `yaml:"containerPort" json:"containerPort"`
	Protocol      string `yaml:"protocol" json:"protocol"`
	HostIP        string `yaml:"hostIP" json:"hostIP"`
}

// 参考Probe
// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#probe-v1-core
type ContainerProbe struct {
	// HTTPGet specifies the http request to perform.
	HttpGet struct {
		Path   string `yaml:"path" json:"path"`
		Port   int    `yaml:"port" json:"port"`
		Host   string `yaml:"host" json:"host"`
		Scheme string `yaml:"scheme" json:"scheme"`
	} `yaml:"httpGet" json:"httpGet"`
	// Number of seconds after the container has started before liveness probes are initiated.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	InitialDelaySeconds int `yaml:"initialDelaySeconds" json:"initialDelaySeconds"`

	// Number of seconds after which the probe times out. Defaults to 1 second. Minimum value is 1.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	TimeoutSeconds int `yaml:"timeoutSeconds" json:"timeoutSeconds"`

	// How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1.
	PeriodSeconds int `yaml:"periodSeconds" json:"periodSeconds"`
}

// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#envvar-v1-core
type EnvVar struct {
	Name  string `yaml:"name" json:"name"`
	Value string `yaml:"value" json:"value"`
}

// 看文档：VolumeMount用在Container中
// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#volumemount-v1-core
type VolumeMount struct {
	Name      string `yaml:"name" json:"name"`
	MountPath string `yaml:"mountPath" json:"mountPath"`
	ReadOnly  bool   `yaml:"readOnly" json:"readOnly"`
}

type LifecycleHandler struct {
	Exec struct {
		Command []string `yaml:"cmd,flow"`
	} `yaml:"exec"`
}

// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#lifecycle-v1-core
type ContainerLifecycle struct {
	PostStart LifecycleHandler `yaml:"postStart"`
	PreStop   LifecycleHandler `yaml:"preStop"`
}

// 关于CPU和Memory怎么写，看这里
// https://kubernetes.io/zh-cn/docs/concepts/configuration/manage-resources-containers/
type ContainerResourcesTypes struct {
	CPU    int `yaml:"cpu" json:"cpu"`       // 代表CPU的占比，最大是10^9
	Memory int `yaml:"memory" json:"memory"` // 代表内存的占比，单位是byte
}

// 这个当你为 Pod 中的 Container 指定了资源 请求时， kube-scheduler 就利用该信息决定将 Pod 调度到哪个节点上。
// 当你还为 Container 指定了资源 限制 时，kubelet 就可以确保运行的容器不会使用超出所设限制的资源。
//
// kubelet 还会为容器预留所 请求 数量的系统资源，供其使用。
//
// https://kubernetes.io/zh-cn/docs/concepts/configuration/manage-resources-containers/
type ContainerResources struct {
	Limits   ContainerResourcesTypes `yaml:"limits"`   // 限制的资源(目前主要是CPU和Memory)
	Requests ContainerResourcesTypes `yaml:"requests"` // 请求的资源
}

// Conatiner结构体
// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#container-v1-core
type Container struct {
	// Name代表容器的名字
	Name string `yaml:"name" json:"name"`

	// Image代表容器的镜像
	Image string `yaml:"image" json:"image"`

	// ImagePullPolicy代表容器的镜像拉取策略
	ImagePullPolicy string `yaml:"imagePullPolicy" json:"imagePullPolicy" default:"IfNotPresent"`

	// Command代表容器的命令
	Command []string `yaml:"command" json:"command"`

	// Args代表容器的命令行参数
	Args []string `yaml:"args" json:"args"`

	// 容器的环境变量
	Env []EnvVar `yaml:"env"`

	// 容器的资源相关的东西，不能更新，详细看文档
	// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#resourcerequirements-v1-core
	// Compute Resources required by this container. Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	Resources ContainerResources `yaml:"resources"`
	Ports     []ContainerPort    `yaml:"ports"`

	// Periodic probe of container liveness. Container will be restarted if the probe fails.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	LivenessProbe ContainerProbe `yaml:"livenessProbe"`

	// 生命周期相关的命令，主要都是针对容器的启动或者挂了的时候执行的命令
	Lifecycle ContainerLifecycle `yaml:"lifecycle" json:"lifecycle"`

	// 挂载的文件系统的东西
	VolumeMounts []VolumeMount `yaml:"volumeMounts" json:"volumeMounts"`

	// 是否开启TTY
	TTY bool `yaml:"tty" json:"tty" default:"false"`
}

// 参考hostPath
// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#hostpathvolumesource-v1-core
type HostPath struct {
	// 主机上面的目录、文件、Socket甚至都行
	Path string `json:"path" yaml:"path"`

	// Type有下面的取值：参考官方文档
	// https://kubernetes.io/zh-cn/docs/concepts/storage/volumes/#hostpath
	Type string `json:"type" yaml:"type"`
}

// 参考Volume的官方
// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#volume-v1-core
type Volume struct {
	Name     string   `json:"name" yaml:"name"`
	HostPath HostPath `json:"hostPath" yaml:"hostPath"`
}

// 参考Kubernetes API文档
// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#podspec-v1-core
type PodSpec struct {
	// 参考https://kubernetes.io/zh-cn/docs/concepts/workloads/pods/pod-lifecycle/#restart-policy
	// 其可能取值包括 Always、OnFailure 和 Never。默认值是 Always。
	// restartPolicy 适用于 Pod 中的所有容器。restartPolicy 仅针对同一节点上 kubelet 的容器重启动作。
	// 当 Pod 中的容器退出时，kubelet 会按指数回退方式计算重启的延迟（10s、20s、40s、...），其最长延迟
	// 为 5 分钟。 一旦某容器执行了 10 分钟并且没有出现问题，kubelet 对该容器的重启回退计时器执行重置操作。
	RestartPolicy string `json:"restartPolicy" yaml:"restartPolicy" default:"Always"`

	// 如果指定了nodeName，那么Pod将会被调度到指定的节点上
	NodeName string `json:"nodeName" yaml:"nodeName"`

	// 容器的集合
	Containers []Container `json:"containers" yaml:"containers"`

	// 一个键值对的map，用来给Pod打标签
	NodeSelector map[string]string `json:"nodeSelector" yaml:"nodeSelector"`

	// pod的挂载的文件系统的东西
	Volumes []Volume `json:"volumes" yaml:"volumes"`
}

type Pod struct {
	Basic `yaml:",inline"`
	Spec  PodSpec `json:"spec" yaml:"spec"`
}

// PodStatus是用来存储Pod的状态的，同时也存储了Pod的一些元数据
// type ContainerStatus

// Pod的Phase
const (
	// PodPending代表Pod处于Pending状态
	PodPending     = "Pending"
	PodRunning     = "Running"
	PodSucceeded   = "Succeeded"
	PodFailed      = "Failed"
	PodUnknown     = "Unknown"
	PodTerminating = "Terminating"
)

// PodStatus是用来存储Pod的状态的
// 参考官方文档：https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#podstatus-v1-core
type PodStatus struct {
	// IP address allocated to the pod. Routable at least within the cluster. Empty if not yet allocated.
	PodIP string `json:"podIP" yaml:"podIP"`

	// Phase参考官方文档
	// https://kubernetes.io/zh-cn/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase
	// Pending（悬决）   Pod 已被 Kubernetes 系统接受，但有一个或者多个容器尚未创建亦未运行。
	//                  此阶段包括等待 Pod 被调度的时间和通过网络下载镜像的时间。
	// Running（运行中） Pod 已经绑定到了某个节点，Pod 中所有的容器都已被创建。至少有一个容器仍在运行，
	//                   或者正处于启动或重启状态。
	// Succeeded（成功） Pod 中的所有容器都已成功终止，并且不会再重启。
	// Failed（失败）	 Pod 中的所有容器都已终止，并且至少有一个容器是因为失败终止。也就是说，容器以
	//                  非 0 状态退出或者被系统终止。
	// Unknown（未知）	 因为某些原因无法取得 Pod 的状态。这种情况通常是因为与 Pod 所在主机通信失败。
	// Terminating（需要终止） Pod 已被请求终止，但是该终止请求还没有被发送到底层容器。Pod 仍然在运行。
	Phase string `json:"phase" yaml:"phase"`

	// 容器的状态数组
	ContainerStatuses []types.ContainerState `json:"containerStatuses" yaml:"containerStatuses"`

	// 最新的更新时间
	// UpdateTime string `json:"lastUpdateTime" yaml:"lastUpdateTime"`
	UpdateTime time.Time `json:"lastUpdateTime" yaml:"lastUpdateTime"`

	// Pod的容器的资源使用情况
	CpuPercent float64 `json:"cpuPercent" yaml:"cpuPercent"`
	MemPercent float64 `json:"memPercent" yaml:"memPercent"`

}

// PodStore是用来存储Pod的设定和他的状态的
type PodStore struct {
	Basic `yaml:",inline"`
	Spec  PodSpec `json:"spec" yaml:"spec"`
	// Pod的状态
	Status PodStatus `json:"status" yaml:"status"`
}

// 定义Pod到PodStore的转换器
func (p *Pod) ToStore() *PodStore {
	return &PodStore{
		Basic:  p.Basic,
		Spec:   p.Spec,
		Status: PodStatus{},
	}
}

// 定义PodStore到Pod的转换器
func (p *PodStore) ToPod() *Pod {
	return &Pod{
		Basic: p.Basic,
		Spec:  p.Spec,
	}
}

func (p *Pod) GetPodUUID() string {
	return p.Metadata.UUID
}

// 工具函数，用来获取Pod的名字
func (p *PodStore) GetPodName() string {
	return p.Metadata.Name
}

func (p *PodStore) GetPodNamespace() string {
	return p.Metadata.Namespace
}

func (p *PodStore) GetPodUUID() string {
	return p.Metadata.UUID
}

// 以下函数用来实现apiObject.Object接口
func (p *Pod) GetObjectKind() string {
	return p.Kind
}

func (p *Pod) GetObjectName() string {
	return p.Metadata.Name
}

func (p *Pod) GetObjectNamespace() string {
	return p.Metadata.Namespace
}
