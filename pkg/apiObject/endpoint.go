package apiObject


type EndpointSubset struct{
	IP string `yaml:"ip"`
	Port string `yaml:"port"`
}


type Endpoint struct {
	Basic `json:",inline" yaml:",inline"`
	Subsets  []EndpointSubset `yaml:"subsets"`
}