package proxy

import "os"

var NginxPodYamlPath = os.Getenv("MINIK8S_PATH") + "util/nginx/yaml/dns-nginx-pod.yaml"
var NginxServiceYamlPath = os.Getenv("MINIK8S_PATH") + "util/nginx/yaml/dns-nginx-service.yaml"
var NginxDnsYamlPath = os.Getenv("MINIK8S_PATH") + "util/nginx/yaml/dns-nginx-dns.yaml"
