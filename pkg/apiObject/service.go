package apiObject

type ServiceStatus struct {
	//
	Endpoints []Endpoint
	Phase     string
}

// ServicePort contains information on service's port.
type ServicePort struct {
	// AppProtocol    string   `yaml:"appProtocol"`
	Name string `yaml:"name"`
	// The port that will be exposed by this service.
	Port int `yaml:"port"`
	// The port on each node on which this service is exposed when type is NodePort or LoadBalancer.
	NodePort int `yaml:"nodePort"`

	TargetPort int `yaml:"targetPort"`

	// The IP protocol for this port. Supports "TCP", "UDP", and "SCTP". Default is TCP.
	Protocol string `yaml:"protocol"`
}

// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#servicespec-v1-core
type ServiceSpec struct {
	// 将service流量路由到具有与此selector匹配的标签键和值的pod。
	Selector map[string]string `yaml:"selector"`
	// 该service所暴露的端口列表。
	Ports []ServicePort `yaml:"ports"`
	// 是否将为LoadBalancer类型的service自动分配节点端口。默认为 "true"。
	AllocateLoadBalancerNodePorts bool `yaml:"allocateLoadBalancerNodePorts"`
	// 决定该service的暴露方式。默认为 ClusterIP。有效选项是ExternalName、ClusterIP、NodePort和LoadBalancer。
	Type string `yaml:"type"`
	// service的IP地址，通常是随机分配的。
	ClusterIP string `yaml:"clusterIP"`
	// 分配给该服务的IP地址列表，通常是随机分配的。
	ClusterIPs []string `yaml:"clusterIPs"`
}

type Service struct {
	Basic `json:",inline" yaml:",inline"`
	Spec  ServiceSpec `json:"spec" yaml:"spec"`
}

// ServiceStore用来存储Service的设定和它的状态
type ServiceStore struct {
	Basic `json:",inline" yaml:",inline"`
	Spec  ServiceSpec `json:"spec" yaml:"spec"`
	// Service的状态
	Status ServiceStatus `json:"status" yaml:"status"`
}

// 定义Service到ServiceStore的转换器
func (s *Service) ToServiceStore() *ServiceStore {
	return &ServiceStore{
		Basic:  s.Basic,
		Spec:   s.Spec,
		Status: ServiceStatus{},
	}
}

// 定义ServiceStore到Service的转换器
func (s *ServiceStore) ToService() *Service {
	return &Service{
		Basic: s.Basic,
		Spec:  s.Spec,
	}
}

func (s *Service) GetAPIVersion() string {
	return s.Basic.APIVersion
}

func (s *Service) GetKind() string {
	return s.Basic.Kind
}

func (s *Service) GetType() string {
	return s.Spec.Type
}

func (s *Service) GetPorts() []ServicePort {
	return s.Spec.Ports
}

func (s *Service) GetName() string {
	return s.Basic.Metadata.Name
}

func (s *Service) GetNamespace() string {
	return s.Basic.Metadata.Namespace
}

func (s *ServiceStore) GetAPIVersion() string {
	return s.Basic.APIVersion
}

func (s *ServiceStore) GetKind() string {
	return s.Basic.Kind
}

func (s *ServiceStore) GetType() string {
	return s.Spec.Type
}

func (s *ServiceStore) GetPorts() []ServicePort {
	return s.Spec.Ports
}

func (s *ServiceStore) GetName() string {
	return s.Basic.Metadata.Name
}

func (s *ServiceStore) GetNamespace() string {
	return s.Basic.Metadata.Namespace
}

// 以下函数用来是实现apiObject.Object接口
func (s *Service) GetObjectKind() string {
	return s.Kind
}

func (s *Service) GetObjectName() string {
	return s.Metadata.Name
}

func (s *Service) GetObjectNamespace() string {
	return s.Metadata.Namespace
}

// 从Condition到ServiceStatus基本都是和负载均衡相关的内容，目前暂不考虑实现

// type Condition struct {
// 	UpdateTime string `yaml:"lastUpdateTime"`
// 	// 描述status转换的原因
// 	Message string `yaml:"message"`
// }

// type PortStatus struct {
// 	Port int `yaml:"port"`
// 	Protocol  string `yaml:"protocol"`
// }

// // LoadBalancerIngress表示负载均衡器ingress point的状态：用于该service的流量应该被发送到一个ingress point。
// type LoadBalancerIngress struct {
// 	HostName     string  `yaml:"hostname"`
// 	IP			 string `yaml:"ip"`
// 	Ports   []PortStatus  `yaml:"ports"`
// }

// type LoadBalancerStatus struct {
// 	Ingress   []LoadBalancerIngress `yaml:"ingress"`
// }

// // 参考官方文档：https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#servicestatus-v1-core
// type ServiceStatus struct {
// 	// 当前Serive的状态，由Condition类的array组成, 暂时用不上
// 	Condition []Condition  `yaml:"conditions" json:"conditions"`
// 	// LoadBalancer的状态，暂时用不上
// 	LoadBalancer LoadBalancerStatus `yaml:"loadBalancer" json:"loadBalancer"`
// }
