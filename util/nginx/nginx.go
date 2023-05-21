package nginx

import (
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	"os"
)

const (
	nginxListenPort = 80
)

func FormatConf(dns apiObject.Dns) string {
	commentStr := fmt.Sprintf("# %s.conf\n", dns.Spec.Host)
	formatStr := fmt.Sprintf("server {\n\tlisten %d;\n\tserver_name %s;\n", nginxListenPort, dns.Spec.Host)
	formatStr = commentStr + formatStr
	locationStr := "\tlocation %s {\n\t\tproxy_pass http://%s:%s/;\n\t}\n"
	for _, p := range dns.Spec.Paths {
		path := p.SubPath
		if path[0] != '/' {
			path = "/" + path
		}
		serviceIP := p.SvcIp
		servicePort := p.SvcPort
		formatStr += fmt.Sprintf(locationStr, path, serviceIP, servicePort)
	}
	formatStr += "}"
	return formatStr
}

func WriteConf(dns apiObject.Dns, conf string) error {
    // 将配置文件写入到nginx的配置文件中
    confFileName := fmt.Sprintf("%s.conf", dns.Spec.Host)
    confFilePath := fmt.Sprintf(config.NginxConfigPath + confFileName)
    file, err := os.Create(confFilePath) // 使用 os.Create() 函数打开文件以进行写入
    if err != nil {
        k8log.ErrorLog("nginx", "WriteConf: create file failed "+err.Error())
        return err
    }
	
    // 将文件截断为空或创建一个新文件
    err = file.Truncate(0)
    if err != nil {
		k8log.ErrorLog("nginx", "WriteConf: truncate file failed "+err.Error())
        return err
    }
	
    _, err = file.Write([]byte(conf)) // 将字符串写入文件
    if err != nil {
		k8log.ErrorLog("nginx", "WriteConf: write file failed "+err.Error())
        return err
    }
	defer file.Close()

    return nil
}