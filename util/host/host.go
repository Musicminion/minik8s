package host

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// 特别感谢下面的库，提供了很多系统信息的获取

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
		if name == "ens33" || name == "eth0" || name == "ens3"{
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
	return "", errors.New("no interface or no named ens33/ens3/eth0")
}

// GetSystemMemoryUsage 返回当前系统内存使用率
func GetHostSystemMemoryUsage() (float64, error) {
	cmd := exec.Command("ps", "axm", "-o", "%mem")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	memUsage := 0.0
	lines := strings.Split(string(output), "\n")
	for _, line := range lines[1 : len(lines)-1] {
		line = strings.TrimSpace(line)
		if line == "" || line == "-" {
			continue
		}
		usage, err := strconv.ParseFloat(line, 64)
		if err != nil {
			return 0, err
		}
		memUsage += usage
	}

	return memUsage, nil
}

// GetSystemCPUUsage 返回当前系统CPU使用率
func GetHostSystemCPUUsage() (float64, error) {
	cmd := exec.Command("ps", "-A", "-o", "%cpu")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	cpuUsage := 0.0
	lines := strings.Split(string(output), "\n")
	for _, line := range lines[1 : len(lines)-1] {
		usage, err := strconv.ParseFloat(strings.TrimSpace(line), 64)
		if err != nil {
			return 0, err
		}
		cpuUsage += usage
	}

	return cpuUsage, nil
}
