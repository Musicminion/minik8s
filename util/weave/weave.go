package weave

import (
	"errors"
	"miniK8s/pkg/k8log"
	"os/exec"
	"regexp"
)

//TODO: weave的ip返回值

func WeaveAttach1(id string) (string, error) {
	out, err := exec.Command("weave", "attach", id).Output()
	if err != nil {
		k8log.ErrorLog("Weave_util", "weave attach err: "+err.Error()+string(out))
		return "", err
	}
	k8log.InfoLog("Weave_util", "weave attach out: "+string(out))
	// 将字节数组转换为字符串
	output := string(out)
	// 在字符串中查找IP地址
	re := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	matches := re.FindStringSubmatch(output)
	if len(matches) < 2 {
		return "", errors.New("could not find IP address in command output")
	}
	// 返回第一个匹配项（IP地址）
	return matches[1], nil
}

func WeaveAttach(containerID, ip string) (string, error) {
	if containerID == "" {
		return "", errors.New("containerID is empty")
	}

	k8log.DebugLog("Weave_util", "weave attach id: "+containerID+" ip: "+ip)

	if ip == "" {
		out, err := exec.Command("weave", "attach", containerID).Output()
		if err != nil {
			return string(out), err
		}
		return string(out), err
	}

	if out, err := exec.Command("weave", "attach", ip, containerID).Output(); err != nil {
		k8log.DebugLog("Weave_util", "weave attch err: "+err.Error()+string(out))
		return string(out), err
	} else {
		k8log.DebugLog("Weave_util", "weave attch out: "+string(out))
		return string(out), err
	}
}
