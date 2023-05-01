package weave

import (
	"miniK8s/pkg/k8log"
	"os/exec"
)

func WeaveAttach(id, ip string) error {
	if out, err := exec.Command("weave", "attach", ip, id).Output(); err != nil {
		k8log.ErrorLog("Weave_util", "weave attch err: "+err.Error()+string(out))
		return err
	} else {
		k8log.InfoLog("Weave_util", "weave attch out: "+string(out))
		return nil
	}
}
