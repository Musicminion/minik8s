package weave

import (
	"errors"
	"miniK8s/pkg/k8log"
	"os/exec"
	"regexp"
)

func WeaveAttach(containerID string) (string, error) {
	if containerID == "" {
		return "", errors.New("containerID is empty")
	}

	out, err := exec.Command("weave", "attach", containerID).Output()
	if err != nil {
		k8log.DebugLog("Weave_util", "weave attach err: "+err.Error()+string(out))
		return string(out), err
	}
	return string(out), err
}

func WeaveFindIpByContainerID(containerID string) (string, error) {
	if containerID == "" {
		return "", errors.New("containerID is empty")
	}

	out, err := exec.Command("weave", "ps", containerID).Output()

	if err != nil {
		k8log.ErrorLog("Weave_util", "weave ps err: "+err.Error()+string(out))
		return "", err
	}

	// k8log.DebugLog("Weave_util", "weave ps out: "+string(out))

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
