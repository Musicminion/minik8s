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

type IptableManager interface {
	CreateService(serviceUpdate *entity.ServiceUpdate)
}

type iptableManager struct {
	// SvcChain     map[string]map[string]
}

var ipt *iptables.IPTables

func New() IptableManager {
	return &iptableManager{}
}

func (im *iptableManager) CreateService(serviceUpdate *entity.ServiceUpdate) {

}

func (im *iptableManager) DeleteService(serviceUpdate *entity.ServiceUpdate) {

}

func (im *iptableManager) UpdateService(serviceUpdate *entity.ServiceUpdate) {

}

func init_iptables() {
	// 创建 iptables 的实例
	ipt, _ = iptables.New()

	// 删除旧规则，设置 NAT 表的策略
	ipt.ClearChain("nat", "PREROUTING")
	ipt.ClearChain("nat", "INPUT")
	ipt.ClearChain("nat", "OUTPUT")
	ipt.ClearChain("nat", "POSTROUTING")
	ipt.ChangePolicy("nat", "PREROUTING", "ACCEPT")
	ipt.ChangePolicy("nat", "INPUT", "ACCEPT")
	ipt.ChangePolicy("nat", "OUTPUT", "ACCEPT")
	ipt.ChangePolicy("nat", "POSTROUTING", "ACCEPT")

	// 创建 NAT 表中的新链
	ipt.NewChain("nat", "KUBE-SERVICES")
	ipt.NewChain("nat", "KUBE-POSTROUTING")
	ipt.NewChain("nat", "KUBE-MARK-MASQ")
	ipt.NewChain("nat", "KUBE-NODEPORTS")

	// 往 NAT 表中的链中添加规则
	ipt.Append("nat", "PREROUTING", "-j KUBE-SERVICES", "-m comment --comment \"kubernetes service portals\"")
	ipt.Append("nat", "OUTPUT", "-j KUBE-SERVICES", "-m comment --comment \"kubernetes service portals\"")
	ipt.Append("nat", "POSTROUTING", "-j KUBE-POSTROUTING", "-m comment --comment \"kubernetes postrouting rules\"")

	ipt.Insert("nat", "KUBE-MARK-MASQ", 1, "-j MARK", "--set-xmark 0x4000/0x4000")
	ipt.Insert("nat", "KUBE-POSTROUTING", 1, "-m comment --comment \"kubernetes service traffic requiring SNAT\"", "-j MASQUERADE", "-m mark --mark 0x4000/0x4000")

}

func create_service_chain(serviceUpdate *entity.ServiceUpdate) {
	clusterIp := serviceUpdate.ServiceTarget.Service.Spec.ClusterIP
	seviceName := serviceUpdate.ServiceTarget.Service.Metadata.Name
	ports := serviceUpdate.ServiceTarget.Service.Spec.Ports
	var pod_ip_list []string
	for _, endpoint := range serviceUpdate.ServiceTarget.Endpoints {
		pod_ip_list = append(pod_ip_list, endpoint.IP)
	}

	for _, eachports := range ports {
		port := eachports.Port
		protocol := eachports.Protocol
		targetPort := eachports.TargetPort
		set_iptables_clusterIp(seviceName, clusterIp, port, protocol, targetPort, []string{})
	}

}

func set_iptables_clusterIp(serviceName string, clusterIP string, port int, protocol string, targetPort int32, podIPList []string) {
	if ipt == nil {
		k8log.DebugLog("KUBEPROXY", "iptables is nil")
		return
	}

	if _, err := ipt.Exists("nat", "KUBE-SVC-"+uuid.NewUUID()); err != nil {
		k8log.DebugLog("KUBEPROXY", "Failed to check the existence of kubesvc chain: "+err.Error())
	}
	if _, err := ipt.Exists("nat", "KUBE-SEP-"+uuid.NewUUID()); err != nil {
		k8log.DebugLog("KUBEPROXY", "Failed to check the existence of kubesep chain: "+err.Error())
	}

	// 添加 NAT 链
	if err := ipt.NewChain("nat", "KUBE-SVC-"+uuid.NewUUID()); err != nil {
		k8log.DebugLog("KUBEPROXY", "Failed to create kubesvc chain: "+err.Error())
	}
	// 添加 NAT 规则，重定向流量到服务的集群 IP
	if err := ipt.Insert("nat", "KUBE-SERVICES", 1, "-m", "comment", "--comment",
		serviceName+": cluster IP", "-p", protocol, "--dport", string(rune(port)),
		"-m", protocol, "--destination", clusterIP+"/"+ strconv.Itoa(config.IP_PREFIX_LENGTH), "-j", "KUBE-SVC-"+uuid.NewUUID()); err != nil {
		k8log.DebugLog("KUBEPROXY", "Failed to insert KUBE-SERVICES rule for kubesvc chain: "+err.Error())
	}
	// 添加 NAT 规则，标记流量为 MASQUERADE
	if err := ipt.Insert("nat", "KUBE-SERVICES", 1, "-m", "comment", "--comment",
		serviceName+": cluster IP", "-p", protocol, "--dport", strconv.Itoa(port),
		"-j", "KUBE-MARK-MASQ", "-m", protocol, "--destination", clusterIP+"/"+strconv.Itoa(config.IP_PREFIX_LENGTH)); err != nil {
		k8log.DebugLog("KUBEPROXY", "Failed to insert KUBE-SERVICES rule for KUBE-MARK-MASQ chain: "+err.Error())
	}

	podNum := len(podIPList)
	for i := podNum - 1; i >= 0; i-- {
		// 为每个pod创建一个KUBE-SEP-UUID 的chain
		if err := ipt.NewChain("nat", "KUBE-SEP-"+uuid.NewUUID()); err != nil {
			k8log.DebugLog("KUBEPROXY", "Failed to create kubesep chain: "+err.Error())
		}
		// 使用随机策略，将流量随机重定向到某个 Pod
		prob := 1 / (podNum - i)
		if i == podNum-1 { // 在最后一个 Pod 上，直接将流量重定向到 KUBE-SEP-UUID 链
			if err := ipt.Insert("nat", "KUBE-SVC-", 1, "-j", "KUBE-SEP-"+uuid.NewUUID()); err != nil {
				k8log.DebugLog("KUBEPROXY", "Failed to create kubesvc chain: "+err.Error())
			}
		} else { // 使用 iptables 的随机策略，将流量随机重定向到某个 Pod
			if err := ipt.Insert("nat", "KUBE-SVC-", 1, "-j", "KUBE-SEP-"+uuid.NewUUID(),
				"-m", "statistic", "--mode", "random", "--probability", strconv.Itoa(prob)); err != nil {
				k8log.DebugLog("KUBEPROXY", "Failed to create kubesvc chain: "+err.Error())
			}
		}
		// 将流量 DNAT 到 Pod IP 和端口
		if err := ipt.Insert("nat", "KUBE-SEP-"+uuid.NewUUID(), 1, "-j", "DNAT",
			"-p", protocol, "-d", podIPList[i], "--dport", string(targetPort),
			"-m", protocol, "--to-destination", podIPList[i]+":"+string(targetPort)); err != nil {
			k8log.DebugLog("KUBEPROXY", "Failed to create kubesvc chain: "+err.Error())
		}
		// 将源 IP 地址标记为 NAT
		if err := ipt.Insert("nat", "KUBE-SEP-"+uuid.NewUUID(), 1, "-j", "KUBE-MARK-MASQ",
			"-s", podIPList[i]+"/"+strconv.Itoa(config.IP_PREFIX_LENGTH)); err != nil {
			k8log.DebugLog("KUBEPROXY", "Failed to create kubesvc chain: "+err.Error())
		}

	}
}

func SaveIPTables(path string) error {
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


func RestoreIPTables(path string) error {
    cmd := exec.Command("iptables-restore", "-c", path)
    if out, err := cmd.CombinedOutput(); err != nil {
        log.Printf("failed to restore iptables: %v, output: %s", err, out)
        return err
    }
    log.Printf("iptables rules have been restored from %v", path)
    return nil
}

func ClearIPTables() {
	ipt.ClearAll()
	log.Printf("iptables rules have been cleared")
}

func Run() {
	init_iptables()
	set_iptables_clusterIp("test", "10.244.10.3", 80, "tcp", 80, []string{"10.1.244.1", "10.1.244.3"})
}