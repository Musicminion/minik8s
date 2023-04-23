package host

import (
	"errors"
	"fmt"
	"net"
	"os"
)

func GetHostName() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}

func GetHostIp() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	// 遍历每个网络接口
	for _, i := range interfaces {
		// 获取网络接口的名称
		name := i.Name
		if name == "ens33" {
			// 获取网络接口的地址信息
			addrs, err := i.Addrs()
			if err != nil {
				fmt.Println("Error:", err)
				return "", err
			}

			// 遍历每个地址
			for _, addr := range addrs {
				// 检查地址的类型是否为IP地址
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
					// 获取IP地址
					return ipnet.IP.String(), nil
				}
			}
		}
	}
	return "", errors.New("no interface or no named ens33")
}
