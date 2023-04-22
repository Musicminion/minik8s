package apiObject

type Node struct {
	Basic `yaml:",inline"`
	IP    string `json:"ip" yaml:"ip"`
}
