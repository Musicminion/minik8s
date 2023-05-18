package apiObject

// type EndpointSubset struct{
// 	IP string `yaml:"ip"`
// 	Port string `yaml:"port"`
// }

type Endpoint struct {
	Basic   `json:",inline" yaml:",inline"`
	PodUUID string `yaml:"podUUID"`
	IP      string `yaml:"ip"`
	Port    string `yaml:"port"`
}

func (ep *Endpoint) GetUUID() string {
	return ep.Basic.Metadata.UUID
}

func (ep *Endpoint) GetIP() string {
	return ep.IP
}

func (ep *Endpoint) GetPort() string {
	return ep.Port
}

func (ep *Endpoint) SetUUID(uuid string) {
	ep.Basic.Metadata.UUID = uuid
}

func (ep *Endpoint) SetIP(ip string) {
	ep.IP = ip
}

func (ep *Endpoint) SetPort(port string) {
	ep.Port = port
}

func (ep *Endpoint) GetPodUUID() string {
	return ep.PodUUID
}

func (ep *Endpoint) SetPodUUID(podUUID string) {
	ep.PodUUID = podUUID
}
