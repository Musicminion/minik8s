package proxy

import (
	"io/ioutil"
	"log"
	"miniK8s/pkg/config"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/k8log"
	"miniK8s/util/uuid"
	"os/exec"
	"strconv"

	"github.com/coreos/go-iptables/iptables"
)

// type IptableManager interface {
// 	CreateService(serviceUpdate *entity.ServiceUpdate)
// }

type IptableManager struct {
	// SvcChain     map[string]map[string]
	ipt *iptables.IPTables
}

// var ipt *iptables.IPTables

func New() *IptableManager {
	iptableManager := &IptableManager{}
	iptableManager.init_iptables()
	return iptableManager
}

func (im *IptableManager) CreateService(serviceUpdate *entity.ServiceUpdate) {
	// chains, err := ipt.ListChains("nat")
	// if err != nil {
	// 	k8log.ErrorLog("KUBEPROXY", "Failed to list chains: "+err.Error())
	// }
	// for _, chain := range chains {
	// 	// create chain
	// 	if chain == "KUBE-SVC-" + uuid.NewUUID() {
	// 		k8log.ErrorLog("KUBEPROXY", "chain already exists")
	// 	}
	// 	else {
	// 		if err := ipt.NewChain("nat", "KUBE-SVC-"+uuid.NewUUID()); err != nil {
	// 			k8log.ErrorLog("KUBEPROXY", "Failed to create kubesvc chain: "+err.Error())
	// 		}
	// 	}
	// }
	im.create_service_chain(serviceUpdate)
}

func (im *IptableManager) DeleteService(serviceUpdate *entity.ServiceUpdate) {

}

func (im *IptableManager) UpdateService(serviceUpdate *entity.ServiceUpdate) {

}

func (im *IptableManager) init_iptables() {
	// 创建 iptables 的实例
	im.ipt, _ = iptables.New()

	// 删除旧规则，设置 NAT 表的策略
	im.ipt.ClearChain("nat", "PREROUTING")
	im.ipt.ClearChain("nat", "INPUT")
	im.ipt.ClearChain("nat", "OUTPUT")
	im.ipt.ClearChain("nat", "POSTROUTING")
	im.ipt.ChangePolicy("nat", "PREROUTING", "ACCEPT")
	im.ipt.ChangePolicy("nat", "INPUT", "ACCEPT")
	im.ipt.ChangePolicy("nat", "OUTPUT", "ACCEPT")
	im.ipt.ChangePolicy("nat", "POSTROUTING", "ACCEPT")

	// 创建 NAT 表中的新链
	im.ipt.NewChain("nat", "KUBE-SERVICES")
	im.ipt.NewChain("nat", "KUBE-POSTROUTING")
	im.ipt.NewChain("nat", "KUBE-MARK-MASQ")
	im.ipt.NewChain("nat", "KUBE-NODEPORTS")

	// 往 NAT 表中的链中添加规则
	im.ipt.Append("nat", "PREROUTING", "-j KUBE-SERVICES", "-m comment --comment \"kubernetes service portals\"")
	im.ipt.Append("nat", "OUTPUT", "-j KUBE-SERVICES", "-m comment --comment \"kubernetes service portals\"")
	im.ipt.Append("nat", "POSTROUTING", "-j KUBE-POSTROUTING", "-m comment --comment \"kubernetes postrouting rules\"")

	im.ipt.Insert("nat", "KUBE-MARK-MASQ", 1, "-j MARK", "--set-xmark 0x4000/0x4000")
	im.ipt.Insert("nat", "KUBE-POSTROUTING", 1, "-m comment --comment \"kubernetes service traffic requiring SNAT\"", "-j MASQUERADE", "-m mark --mark 0x4000/0x4000")

}

func (im *IptableManager) create_service_chain(serviceUpdate *entity.ServiceUpdate) {
	clusterIp := serviceUpdate.ServiceTarget.Service.Spec.ClusterIP
	seviceName := serviceUpdate.ServiceTarget.Service.Metadata.Name
	ports := serviceUpdate.ServiceTarget.Service.Spec.Ports
	var pod_ip_list []string
	for _, endpoint := range serviceUpdate.ServiceTarget.Endpoints {
		pod_ip_list = append(pod_ip_list, endpoint.IP)
	}

	for _, eachports := range ports {
		k8log.DebugLog("KUBEPROXY", "port: "+strconv.Itoa(eachports.Port))
		port := eachports.Port
		protocol := eachports.Protocol
		targetPort := eachports.TargetPort
		im.set_iptables_clusterIp(seviceName, clusterIp, port, protocol, targetPort, pod_ip_list)
	}

}

func (im *IptableManager) set_iptables_clusterIp(serviceName string, clusterIP string, port int, protocol string, targetPort int32, podIPList []string) {
	if im.ipt == nil {
		k8log.ErrorLog("KUBEPROXY", "im.iptables is nil")
		return
	}

	if _, err := im.ipt.Exists("nat", "KUBE-SVC-" + uuid.NewUUID()); err != nil {
		k8log.ErrorLog("KUBEPROXY", "Failed to check the existence of kubesvc chain: "+ err.Error())
	}
	if _, err := im.ipt.Exists("nat", "KUBE-SEP-" + uuid.NewUUID()); err != nil {
		k8log.ErrorLog("KUBEPROXY", "Failed to check the existence of kubesep chain: "+ err.Error())
	}

	// 添加 NAT 链
	if err := im.ipt.NewChain("nat", "KUBE-SVC-" + uuid.NewUUID()); err != nil {
		k8log.ErrorLog("KUBEPROXY", "Failed to create kubesvc chain: "+err.Error())
	}
	// 添加 NAT 规则，重定向流量到服务的集群 IP
	if err := im.ipt.Insert("nat", "KUBE-SERVICES", 1, "-m", "comment", "--comment",
		serviceName + ": cluster IP", "-p", protocol, "--dport", string(rune(port)),
		"-m", protocol, "--destination", clusterIP + "/" + strconv.Itoa(config.IP_PREFIX_LENGTH), "-j", "KUBE-SVC-"+uuid.NewUUID()); err != nil {
		k8log.ErrorLog("KUBEPROXY", "Failed to insert KUBE-SERVICES rule for kubesvc chain: "+err.Error())
	}
	// 添加 NAT 规则，标记流量为 MASQUERADE
	if err := im.ipt.Insert("nat", "KUBE-SERVICES", 1, "-m", "comment", "--comment",
		serviceName+": cluster IP", "-p", protocol, "--dport", strconv.Itoa(port),
		"-j", "KUBE-MARK-MASQ", "-m", protocol, "--destination", clusterIP+"/"+strconv.Itoa(config.IP_PREFIX_LENGTH)); err != nil {
		k8log.ErrorLog("KUBEPROXY", "Failed to insert KUBE-SERVICES rule for KUBE-MARK-MASQ chain: "+err.Error())
	}

	podNum := len(podIPList)
	for i := podNum - 1; i >= 0; i-- {
		// 为每个pod创建一个KUBE-SEP-UUID 的chain
		if err := im.ipt.NewChain("nat", "KUBE-SEP-"+uuid.NewUUID()); err != nil {
			k8log.ErrorLog("KUBEPROXY", "Failed to create kubesep chain: "+ err.Error())
		}
		// 使用随机策略，将流量随机重定向到某个 Pod
		prob := 1 / (podNum - i)
		if i == podNum-1 { // 在最后一个 Pod 上，直接将流量重定向到 KUBE-SEP-UUID 链
			if err := im.ipt.Insert("nat", "KUBE-SVC-", 1, "-j", "KUBE-SEP-"+uuid.NewUUID()); err != nil {
				k8log.ErrorLog("KUBEPROXY", "Failed to create kubesvc chain: "+ err.Error())
			}
		} else { // 使用 im.iptables 的随机策略，将流量随机重定向到某个 Pod
			if err := im.ipt.Insert("nat", "KUBE-SVC-", 1, "-j", "KUBE-SEP-" + uuid.NewUUID(),
				"-m", "statistic", "--mode", "random", "--probability", strconv.Itoa(prob)); err != nil {
				k8log.ErrorLog("KUBEPROXY", "Failed to create kubesvc chain: "+ err.Error())
			}
		}
		// 将流量 DNAT 到 Pod IP 和端口
		if err := im.ipt.Insert("nat", "KUBE-SEP-"+uuid.NewUUID(), 1, "-j", "DNAT",
			"-p", protocol, "-d", podIPList[i], "--dport", string(targetPort),
			"-m", protocol, "--to-destination", podIPList[i]+":" + string(targetPort)); err != nil {
			k8log.ErrorLog("KUBEPROXY", "Failed to create kubesvc chain: "+ err.Error())
		}
		// 将源 IP 地址标记为 NAT
		if err := im.ipt.Insert("nat", "KUBE-SEP-"+uuid.NewUUID(), 1, "-j", "KUBE-MARK-MASQ",
			"-s", podIPList[i]+"/" + strconv.Itoa(config.IP_PREFIX_LENGTH)); err != nil {
			k8log.ErrorLog("KUBEPROXY", "Failed to create kubesvc chain: "+ err.Error())
		}

	}
	k8log.DebugLog("KUBEPROXY", "iptables rules have been set for service: " + serviceName)
	im.SaveIPTables("test-save-iptables")
}

func (im *IptableManager) SaveIPTables(path string) error {
	cmd := exec.Command("iptables-save")
	stdout, err := cmd.StdoutPipe()
	println(stdout)
	if err != nil {
		log.Printf("failed to obtain pipe for iptables-save: %v", err)
		return err
	}
	if err := cmd.Start(); err != nil {
		log.Printf("failed to start iptables-save: %v", err)
		return err
	}
	defer cmd.Process.Kill()

	content, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Printf("failed to read iptables-save output: %v", err)
		return err
	}

	if err := ioutil.WriteFile(path, content, 0644); err != nil {
		log.Printf("failed to write to file %v: %v", path, err)
		return err
	}
	log.Printf("iptables rules have been saved to %v", path)
	return nil
}


func (im *IptableManager) RestoreIPTables(path string) error {
    cmd := exec.Command("iptables-restore", "-c", path)
    if out, err := cmd.CombinedOutput(); err != nil {
        log.Printf("failed to restore iptables: %v, output: %s", err, out)
        return err
    }
    log.Printf("iptables rules have been restored from %v", path)
    return nil
}

func (im *IptableManager) ClearIPTables() {
	im.ipt.ClearAll()
	log.Printf("iptables rules have been cleared")
}

func (im *IptableManager) Run() {
	// im.init_iptables()
	im.set_iptables_clusterIp("test", "10.244.10.3", 80, "tcp", 80, []string{"10.1.244.1", "10.1.244.3"})
}