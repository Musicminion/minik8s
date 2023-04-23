package config

import "strconv"

type KubeletConfig struct {
	IfDebug            bool
	APIServerIP        string
	APIServerPort      int
	APIServerScheme    string
	APIServerURLPrefix string
}

func DefaultKubeletConfig() *KubeletConfig {
	apiserverIP := "127.0.0.1"
	apiserverPort := 8090
	apiserverScheme := "http://"
	apiserverURLPrefix := apiserverScheme + apiserverIP + ":" + strconv.Itoa(apiserverPort)
	return &KubeletConfig{
		IfDebug:            false,
		APIServerIP:        apiserverIP,
		APIServerPort:      apiserverPort,
		APIServerScheme:    apiserverScheme,
		APIServerURLPrefix: apiserverURLPrefix,
	}
}
