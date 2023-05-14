package kubeletconfig

import (
	"miniK8s/pkg/config"
	"miniK8s/pkg/listwatcher"
	"strconv"
)

type KubeletConfig struct {
	IfDebug bool
	// 配置API Server的信息
	APIServerIP        string
	APIServerPort      int
	APIServerScheme    string
	APIServerURLPrefix string
	// ListWatch配置信息
	LWConf *listwatcher.ListwatcherConfig
}

func DefaultKubeletConfig() *KubeletConfig {
	apiserverIP := "127.0.0.1"
	apiserverPort := 8090
	apiserverScheme := "http://"
	apiserverURLPrefix := apiserverScheme + apiserverIP + ":" + strconv.Itoa(apiserverPort)
	lwconf := listwatcher.DefaultListwatcherConfig()

	return &KubeletConfig{
		IfDebug:            false,
		APIServerIP:        apiserverIP,
		APIServerPort:      apiserverPort,
		APIServerScheme:    apiserverScheme,
		APIServerURLPrefix: apiserverURLPrefix,
		LWConf:             lwconf,
	}
}

func ProductionKubeletConfig() *KubeletConfig {
	apiserverIP := config.API_Server_IP
	apiserverPort := config.API_Server_Port
	apiserverScheme := config.API_Server_Scheme
	apiserverURLPrefix := config.API_Server_URL_Prefix
	lwconf := listwatcher.DefaultListwatcherConfig()

	return &KubeletConfig{
		IfDebug:            false,
		APIServerIP:        apiserverIP,
		APIServerPort:      apiserverPort,
		APIServerScheme:    apiserverScheme,
		APIServerURLPrefix: apiserverURLPrefix,
		LWConf:             lwconf,
	}
}
