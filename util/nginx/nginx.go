package nginx

import (
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	"os"
)

const (
	NginxListenPort = 80
	NginxSvcName   = "dns-nginx-service"
)

func FormatConf(dns apiObject.Dns) string {
	commentStr := fmt.Sprintf("# %s.conf\n", dns.Spec.Host)
	formatStr := fmt.Sprintf("server {\n\tlisten %d;\n\tserver_name %s;\n", NginxListenPort, dns.Spec.Host)
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
	k8log.DebugLog("nginx", "WriteConf: conf is "+conf)
	// 将配置文件写入到nginx的配置文件中
	confFileName := fmt.Sprintf("%s.conf", dns.Spec.Host)
	confFilePath := fmt.Sprintf(config.NginxConfigPath + confFileName)
	// 判断文件的目录是否存在
	_, err := os.Stat(config.NginxConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 目录不存在
			// 创建目录
			err = os.MkdirAll(config.NginxConfigPath, os.ModePerm)
			if err != nil {
				k8log.ErrorLog("nginx", "WriteConf: mkdir failed "+err.Error())
				return err
			}
		} else {
			// 其他错误
			k8log.ErrorLog("nginx", "WriteConf: stat dir failed "+err.Error())
			return err
		}
	}

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

func DeleteConf(dns apiObject.Dns) error {
	// 删除配置文件
	confFileName := fmt.Sprintf("%s.conf", dns.Spec.Host)
	confFilePath := fmt.Sprintf(config.NginxConfigPath + confFileName)
	// 判断文件的目录是否存在
	_, err := os.Stat(confFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在
			return nil
		} else {
			// 其他错误
			k8log.ErrorLog("nginx", "DeleteConf: stat file failed "+err.Error())
			return err
		}
	}

	err = os.Remove(confFilePath)
	if err != nil {
		k8log.ErrorLog("nginx", "DeleteConf: remove file failed "+err.Error())
		return err
	}

	return nil
}


