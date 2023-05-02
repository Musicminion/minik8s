package weave

import (
	"errors"
	"miniK8s/pkg/k8log"
	"os/exec"
)

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
