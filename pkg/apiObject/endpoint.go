package apiObject


// type EndpointSubset struct{
// 	IP string `yaml:"ip"`
// 	Port string `yaml:"port"`
// }


type Endpoint struct {
	Basic `json:",inline" yaml:",inline"`
	IP string `yaml:"ip"`
	Port string `yaml:"port"`
}

func GetUUID(ep *Endpoint) string {
	return ep.Basic.Metadata.UUID
}

func GetIP(ep *Endpoint) string {
	return ep.IP
}

func GetPort(ep *Endpoint) string {
	return ep.Port
}

func SetUUID(ep *Endpoint, uuid string) {
	ep.Basic.Metadata.UUID = uuid
}

func SetIP(ep *Endpoint, ip string) {
	ep.IP = ip
}

func SetPort(ep *Endpoint, port string) {
	ep.Port = port
}