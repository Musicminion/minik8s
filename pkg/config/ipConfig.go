package config

import "strconv"

const (
	IP_PREFIX_LENGTH      = 32
	Local_Server_IP       = "127.0.0.1"
	Cluster_Master_IP     = "192.168.1.5"
	API_Server_Port       = 8090
	Serveless_Server_Port = 28080
	API_Server_Scheme     = "http://"
	clusterMode           = true // 是否是集群模式
)

var SERVICE_IP_PREFIX = [2]int{192, 168}

func GetMasterIP() string {
	// 如果是集群环境，那么就使用集群环境的IP地址
	// 如果是单机环境，那么就使用本机的IP地址
	if clusterMode {
		return Cluster_Master_IP
	} else {
		return Local_Server_IP
	}
}

// 如果时localhost模式，返回的是 "http://127.0.0.1:8090"
func GetAPIServerURLPrefix() string {
	return API_Server_Scheme + GetMasterIP() + ":" + strconv.Itoa(API_Server_Port)
}

func GetServelessServerURLPrefix() string {
	return API_Server_Scheme + GetMasterIP() + ":" + strconv.Itoa(Serveless_Server_Port)
}
