package apiObject


type Path struct {
	SubPath string `json:"subPath" yaml:"subPath"`
	SvcName string `json:"svcName" yaml:"svcName"`
	SvcPort string `json:"svcPort" yaml:"svcPort"`
	SvcIp   string `json:"svcIp" yaml:"svcIp"`
}

type DnsSpec struct {
	Host string `json:"host" yaml:"host"`
	Paths []Path `json:"paths" yaml:"paths"`
}


type Dns struct {
	Basic `json:",inline" yaml:",inline"`
	Spec  DnsSpec `json:"spec" yaml:"spec"`
}

