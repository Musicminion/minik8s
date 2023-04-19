package apiObject

type ContainerPort struct {
	Name          string `yaml:"name"`
	HostPort      string `yaml:"hostPort"`
	ContainerPort string `yaml:"containerPort"`
	Protocol      string `yaml:"protocol"`
	HostIP        string `yaml:"hostIP"`
}

type Container struct {
	Name            string                       `yaml:"name"`
	Image           string                       `yaml:"image"`
	ImagePullPolicy string                       `yaml:"imagePullPolicy"`
	Command         []string                     `yaml:"cmd,flow"`
	Args            []string                     `yaml:"args,flow"`
	// Env             []EnvVar                     `yaml:"env"`
	// Resources       ContainerResources           `yaml:"resources"`
	Ports           []ContainerPort              `yaml:"ports"`
	// LivenessProbe   ContainerLivenessProbeConfig `yaml:"livenessProbe"`
	// Lifecycle       ContainerLifecycleConfig     `yaml:"lifecycle"`
	// VolumeMounts    []VolumeMount                `yaml:"volumeMounts"`
	TTY             bool                         `yaml:"tty"`
}

type PodSpec struct {
	RestartPolicy string      `yaml:"restartPolicy"`
	Containers    []Container `yaml:"containers"`
	ClusterIp     string      `yaml:"clusterIp,omitempty"`
}

type PodStatus struct {
	Phase string `json:"phase" yaml:"phase"`
	// IP address allocated to the pod. Routable at least within the cluster
	PodIP string `json:"podIP" yaml:"podIP"`
	//error message
	Err string `json:"err" yaml:"err"`
}


type Pod struct {
	Name   string            `json:"name" yaml:"name"`
	Namespace  string        `json:"namespace" yaml:"namespace"`
	Labels map[string]string `json:"labels" yaml:"labels"`
	UID    string            `json:"uid" yaml:"uid"`
	Spec PodSpec  			 `json:"spec" yaml:"spec"`
	Status PodStatus 		 `json:"status" yaml:"status"`
}