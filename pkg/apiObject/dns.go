package apiObject

import "time"

// Dns的Phase
const (
	DnsPending     = "Pending"
	DnsRunning     = "Running"
	DnsSucceeded   = "Succeeded"
	DnsFailed      = "Failed"
	DnsUnknown     = "Unknown"
	DnsTerminating = "Terminating"
)

type Path struct {
	SubPath string `json:"subPath" yaml:"subPath"`
	SvcName string `json:"svcName" yaml:"svcName"`
	SvcPort string `json:"svcPort" yaml:"svcPort"`
	SvcIp   string `json:"svcIp" yaml:"svcIp"`
}

type DnsSpec struct {
	Host  string `json:"host" yaml:"host"`
	Paths []Path `json:"paths" yaml:"paths"`
}

type Dns struct {
	Basic `json:",inline" yaml:",inline"`
	Spec  DnsSpec `json:"spec" yaml:"spec"`
}

type DnsStatus struct {
	Phase      string    `json:"phase" yaml:"phase"`
	UpdateTime time.Time `yaml:"updateTime" json:"updateTime"`
}

type HpaStore struct {
	Basic  `yaml:",inline" json:",inline"`
	Spec   DnsSpec   `yaml:"spec" json:"spec"`
	Status DnsStatus `yaml:"status" json:"status"`
}

func (d *Dns) ToDnsStore() *HpaStore {
	return &HpaStore{
		Basic: d.Basic,
		Spec:  d.Spec,
	}
}

func (ds *HpaStore) ToDns() *Dns {
	return &Dns{
		Basic: ds.Basic,
		Spec:  ds.Spec,
	}
}

// 以下函数用来实现apiObject.Object接口
func (d *Dns) GetObjectKind() string {
	return d.Kind
}

func (d *Dns) GetObjectName() string {
	return d.Metadata.Name
}

func (d *Dns) GetObjectNamespace() string {
	return d.Metadata.Namespace
}

