package proxy

import (
	"os"
	"time"
)

var (
	NginxPodYamlPath     = os.Getenv("MINIK8S_PATH") + "util/nginx/yaml/dns-nginx-pod.yaml"
	NginxServiceYamlPath = os.Getenv("MINIK8S_PATH") + "util/nginx/yaml/dns-nginx-service.yaml"
	NginxDnsYamlPath     = os.Getenv("MINIK8S_PATH") + "util/nginx/yaml/dns-nginx-dns.yaml"
)

var (
	ReloadConfigDelay    = 0 * time.Second
	ReloadConfigInterval = []time.Duration{15 * time.Second}
	ReloadConfigIfLoop   = true
)
