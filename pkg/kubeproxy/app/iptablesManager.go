package proxy

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"miniK8s/pkg/config"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/k8log"
	"miniK8s/util/stringutil"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-iptables/iptables"
)

// type IptableManager interface {
// 	CreateService(serviceUpdate *entity.ServiceUpdate)
// }

type IptableManager struct {
	// SvcChain     map[string]map[string]
	ipt      *iptables.IPTables
	stragegy string
	// serviceName to clusterIP
	// serviceIPMap map[string]string
	serviceDict map[string]map[string]string
}

// var ipt *iptables.IPTables

const (
	RANDOM   = "random"
	ROUNDDOB = "roundrobin"
)

func New() *IptableManager {
	iptableManager := &IptableManager{
		stragegy: RANDOM,
	}
	iptableManager.init_iptables()
	return iptableManager
}

func (im *IptableManager) CreateService(serviceUpdate *entity.ServiceUpdate) {
	clusterIp := serviceUpdate.ServiceTarget.Spec.ClusterIP
	seviceName := serviceUpdate.ServiceTarget.Metadata.Name
	ports := serviceUpdate.ServiceTarget.Spec.Ports
	var pod_ip_list []string
	for _, endpoint := range serviceUpdate.ServiceTarget.Status.Endpoints {
		pod_ip_list = append(pod_ip_list, endpoint.IP)
	}

	if clusterIp == "" {
		clusterIp, _ = im.allocClusterIP()
	}

	for _, eachports := range ports {
		k8log.DebugLog("KUBEPROXY", "port: "+strconv.Itoa(eachports.Port))
		port := eachports.Port
		protocol := eachports.Protocol
		targetPort := eachports.TargetPort
		im.setIPTablesClusterIp(seviceName, clusterIp, port, protocol, targetPort, pod_ip_list)
	}
}

func (im *IptableManager) DeleteService(serviceUpdate *entity.ServiceUpdate) {

	// 删除 iptables 规则
	im.ipt.Delete("nat", "KUBE-SERVICES", "-m", "comment", "--comment")
}

func (im *IptableManager) UpdateService(serviceUpdate *entity.ServiceUpdate) {
	// clusterIp := serviceUpdate.ServiceTarget.Spec.ClusterIP
	// seviceName := serviceUpdate.ServiceTarget.Metadata.Name
	// ports := serviceUpdate.ServiceTarget.Spec.Ports
	// var pod_ip_list []string
	// for _, endpoint := range serviceUpdate.ServiceTarget.Status.Endpoints {
	// 	pod_ip_list = append(pod_ip_list, endpoint.IP)
	// }

	// if clusterIp == "" {
	// 	clusterIp, _ = im.allocClusterIP()
	// }

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
	// im.ipt.NewChain("nat", "KUBE-NODEPORTS")

	time.Sleep(3 * time.Second)

	// 往 NAT 表中的链中添加规则
	im.ipt.Append("nat", "PREROUTING", "-j", "KUBE-SERVICES", "-m", "comment", "--comment", "kubernetes service portals")
	im.ipt.Append("nat", "OUTPUT", "-j", "KUBE-SERVICES", "-m", "comment", "--comment", "kubernetes service portals")
	im.ipt.Append("nat", "POSTROUTING", "-j", "KUBE-POSTROUTING", "-m", "comment", "--comment", "kubernetes postrouting rules")

	im.ipt.Insert("nat", "KUBE-MARK-MASQ", 1, "-j", "MARK", "--or-mark", "0x4000")
	im.ipt.Insert("nat", "KUBE-POSTROUTING", 1, "-m", "comment", "--comment", "kubernetes service traffic requiring SNAT", "-j", "MASQUERADE", "-m", "mark", "--mark", "0x4000/0x4000")
	k8log.InfoLog("KUBEPROXY", "init iptables success")
	im.SaveIPTables("iptables-save")
}

func (im *IptableManager) allocClusterIP() (string, bool) {

	maxTryTime := 1000 // 最大尝试次数
	ipAllocated := make(map[string]bool)
	ip := ""
	for _, service := range im.serviceDict {
		if service["clusterIP"] != "" {
			ipAllocated[service["clusterIP"]] = true
		}
	}
	source := rand.NewSource(time.Now().UnixNano()) // 以当前时间作为随机数种子
	rng := rand.New(source)
	for maxTryTime > 0 {
		maxTryTime--
		// 前两位ip是指定好的
		num0 := strconv.Itoa(config.SERVICE_IP_PREFIX[0])
		num1 := strconv.Itoa(config.SERVICE_IP_PREFIX[1])
		num2 := strconv.Itoa(config.SERVICE_IP_PREFIX[2])
		// num2 := strconv.Itoa(rng.Intn(256))
		num3 := strconv.Itoa(rng.Intn(256))
		ip = strings.Join([]string{num0, num1, num2, num3}, ".")
		if _, ok := ipAllocated[ip]; !ok {
			break
		}
	}
	if maxTryTime <= 0 {
		log.Fatal("No available service cluster IP address")
		return ip, false
	}
	k8log.DebugLog("KUBEPROXY", "allocClusterIP succeeded, ip is "+ip)
	return ip, true
}

/**
 * @Description: 为每个服务创建 iptables 规则
 * @Param: serviceName 服务名称
 * @Param: clusterIP 服务的集群 IP
 * @Param: port 服务的端口
 * @Param: protocol 服务的协议
 * @Param: targetPort 服务的目标端口
 * @Param: podIPList 服务的 Pod IP 列表
 */
func (im *IptableManager) setIPTablesClusterIp(serviceName string, clusterIP string, port int, protocol string, targetPort int, podIPList []string) {
	k8log.DebugLog("KUBEPROXY", "setIPTablesClusterIp: "+serviceName+" "+clusterIP+" "+strconv.Itoa(port)+" "+protocol+" "+strconv.Itoa(targetPort))
	
	if im.ipt == nil {
		k8log.ErrorLog("KUBEPROXY", "im.iptables is nil")
		return
	}

	if _, err := im.ipt.Exists("nat", "KUBE-SVC-"+stringutil.GenerateRandomStr(12)); err != nil {
		k8log.ErrorLog("KUBEPROXY", "Failed to check the existence of kubesvc chain: "+err.Error())
	}
	if _, err := im.ipt.Exists("nat", "KUBE-SEP-"+stringutil.GenerateRandomStr(12)); err != nil {
		k8log.ErrorLog("KUBEPROXY", "Failed to check the existence of kubesep chain: "+err.Error())
	}

	// 添加 NAT 链
	kubesvc := "KUBE-SVC-" + stringutil.GenerateRandomStr(12)
	if err := im.ipt.NewChain("nat", kubesvc); err != nil {
		k8log.ErrorLog("KUBEPROXY", "Failed to create kubesvc chain: "+err.Error())
	}

	// 添加 NAT 规则，重定向流量到服务的集群 IP
	if err := im.ipt.Insert("nat", "KUBE-SERVICES", 1, "-m", "comment", "--comment",
		serviceName+": cluster IP", "-p", protocol, "--dport", strconv.Itoa(port),
		 "--destination", clusterIP+"/"+strconv.Itoa(config.IP_PREFIX_LENGTH), "-j", kubesvc); err != nil {
		k8log.ErrorLog("KUBEPROXY", "Failed to insert KUBE-SERVICES rule for kubesvc chain: "+err.Error())
	}

// 	if err := im.ipt.Insert("nat", "KUBE-SERVICES", 1, "-m", "comment", "--comment",
// 	serviceName+": cluster IP", "-p", protocol, "--dport", strconv.Itoa(port),
// 	"-m", protocol, "--destination", clusterIP+"/"+strconv.Itoa(config.IP_PREFIX_LENGTH), "-j", kubesvc); err != nil {
// 	k8log.ErrorLog("KUBEPROXY", "Failed to insert KUBE-SERVICES rule for kubesvc chain: "+err.Error())
// }

	// 添加 NAT 规则，标记流量为 MASQUERADE
	if err := im.ipt.Insert("nat", "KUBE-SERVICES", 1, "-m", "comment", "--comment",
		serviceName+": cluster IP", "-p", protocol, "--dport", strconv.Itoa(port),
		"-j", "KUBE-MARK-MASQ", "--destination", clusterIP+"/"+strconv.Itoa(config.IP_PREFIX_LENGTH)); err != nil {
		k8log.ErrorLog("KUBEPROXY", "Failed to insert KUBE-SERVICES rule for KUBE-MARK-MASQ chain: "+err.Error())
	}


// 	if err := im.ipt.Insert("nat", "KUBE-SERVICES", 1, "-m", "comment", "--comment",
// 	serviceName+": cluster IP", "-p", protocol, "--dport", strconv.Itoa(port),
// 	"-j", "KUBE-MARK-MASQ", "-m", protocol, "--destination", clusterIP+"/"+strconv.Itoa(config.IP_PREFIX_LENGTH)); err != nil {
// 	k8log.ErrorLog("KUBEPROXY", "Failed to insert KUBE-SERVICES rule for KUBE-MARK-MASQ chain: "+err.Error())
// }

	podNum := len(podIPList)
	k8log.DebugLog("KUBEPROXY", "podNum is "+strconv.Itoa(podNum))
	for i := podNum - 1; i >= 0; i-- {
		kubesep := "KUBE-SEP-" + stringutil.GenerateRandomStr(12)
		// 为每个pod创建一个KUBE-SEP-UUID 的chain
		if err := im.ipt.NewChain("nat", kubesep); err != nil {
			k8log.ErrorLog("KUBEPROXY", "Failed to create kubesep chain: "+err.Error())
		}

		if im.stragegy == RANDOM {
			prob := 1 / (podNum - i)
			if i == podNum-1 { // 在最后一个 Pod 上，直接将流量重定向到 KUBE-SEP-UUID 链
				if err := im.ipt.Insert("nat", kubesvc, 1, "-j", kubesep); err != nil {
					k8log.ErrorLog("KUBEPROXY", "Failed to create kubesvc chain: "+err.Error())
				}
			} else { // 使用 im.iptables 的随机策略，将流量随机重定向到某个 Pod
				if err := im.ipt.Insert("nat", kubesvc, 1, "-j", kubesep,
					"-m", "statistic", "--mode", "random", "--probability", strconv.Itoa(prob)); err != nil {
					k8log.ErrorLog("KUBEPROXY", "Failed to create kubesvc chain: "+err.Error())
				}
			}
		} else if im.stragegy == ROUNDDOB {
			if i == podNum-1 { // 在最后一个 Pod 上，直接将流量重定向到 KUBE-SEP-UUID 链
				if err := im.ipt.Insert("nat", kubesvc, 1, "-j", kubesep); err != nil {
					k8log.ErrorLog("KUBEPROXY", "Failed to create kubesvc chain: "+err.Error())
				}
			} else { // 使用 roundrobin 策略，将流量重定向到下一个 Pod
				if err := im.ipt.Insert("nat", kubesvc, 1, "-j", kubesep,
					"-m", "statistic", "--mode", "nth", "--every", strconv.Itoa(podNum-i)); err != nil {
					k8log.ErrorLog("KUBEPROXY", "Failed to create kubesvc chain: "+err.Error())
				}
			}
		}

		// 将流量 DNAT 到 Pod IP 和端口
		if err := im.ipt.Insert("nat", kubesep, 1, "-j", "DNAT",
			"-p", protocol,
			 "--to-destination", podIPList[i]+":"+strconv.Itoa(targetPort)); err != nil {
			k8log.ErrorLog("KUBEPROXY", "Failed to create kubesvc chain: "+err.Error())
		}
		// 将源 IP 地址标记为 NAT
		if err := im.ipt.Insert("nat", kubesep, 1, "-j", "KUBE-MARK-MASQ",
			"-s", podIPList[i]+"/"+strconv.Itoa(config.IP_PREFIX_LENGTH)); err != nil {
			k8log.ErrorLog("KUBEPROXY", "Failed to create kubesvc chain: "+err.Error())
		}

	}
	k8log.DebugLog("KUBEPROXY", "iptables rules have been set for service: "+serviceName)
	im.SaveIPTables("test-save-iptables")
}

// SaveIPTables将iptables规则保存到一个文件中。
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

// RestoreIPTables从一个文件中恢复iptables规则
func (im *IptableManager) RestoreIPTables(path string) error {
	cmd := exec.Command("iptables-restore", "-c", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("failed to restore iptables: %v, output: %s", err, out)
		return err
	}
	log.Printf("iptables rules have been restored from %v", path)
	return nil
}

func (im *IptableManager) Run() {
	// im.init_iptables()
	im.setIPTablesClusterIp("test", "10.32.10.3", 80, "tcp", 80, []string{"10.1.32.1", "10.1.32.3"})
}

func (im *IptableManager) DeletePrefix(table, chain, prefix string) error {
	output, err := im.ipt.List(table, chain)
	if err != nil {
		return err
	}
	rules := strings.TrimRight(strings.Join(output, "\n"), "\n")
	for _, rule := range strings.Split(rules, "\n") {
		if strings.HasPrefix(rule, "-A") && strings.Contains(rule, fmt.Sprintf("%s-", prefix)) {
			// extract the rulespec from the rule
			rulespec := strings.Split(rule, " ")
			rulespec = rulespec[1:]
			// remove the rule
			err = im.ipt.Delete(table, chain, rulespec...)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ClearIPTables 清除所有的 iptables 规则
func (im *IptableManager) ClearIPTables() {
	im.ipt.ClearAll()
	im.ipt.DeleteAll()
	// im.ipt.ClearChain("nat", "KUBE-SVC")
	im.ipt.ClearAndDeleteChain("nat", "KUBE-SVC")
	im.DeletePrefix("nat", "KUBE-SERVICES", "KUBE-SVC-")

	log.Printf("iptables rules have been cleared")
}
