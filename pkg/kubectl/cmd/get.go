package cmd

import (
	"errors"
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	netrequest "miniK8s/util/netRequest"
	"miniK8s/util/stringutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Kubectl get can get apiObject in a declarative way",
	Long:  "Kubectl get can get apiObject in a declarative way, usage kubectl get [pod|service|job|deploy|dns]/[pods|services|jobs|deploys]",
	Run:   getHandler,
}

func init() {
	getCmd.PersistentFlags().StringP("namespace", "n", "", "Namespace")
}

type GetObject string

const (
	Get_Kind_Pod        GetObject = "pod"
	Get_Kind_Service    GetObject = "service"
	Get_Kind_Job        GetObject = "job"
	Get_Kind_Replicaset GetObject = "replicaset"
	Get_Kind_Dns        GetObject = "dns"

	Get_Kind_Pods        GetObject = "pods"
	Get_Kind_Services    GetObject = "services"
	Get_Kind_Jobs        GetObject = "jobs"
	Get_Kind_Replicasets GetObject = "replicasets"
)

func getHandler(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("getHandler: no args, please specify [pod|service|job|replicaset]/[pods|services|jobs|deploys]")
		cmd.Usage()
		return
	}

	args[0] = strings.ToLower(args[0])

	switch args[0] {
	case string(Get_Kind_Pod), string(Get_Kind_Pods):
		getPodHandler(cmd, args)
	case string(Get_Kind_Service), string(Get_Kind_Services):
		getServiceHandler(cmd, args)
	case string(Get_Kind_Job), string(Get_Kind_Jobs):
		getJobHandler(cmd, args)
	case string(Get_Kind_Replicaset), string(Get_Kind_Replicasets):
		// fmt.Println("Kind: Deployment")
	case string(Get_Kind_Dns):
		getDnsHandler(cmd, args)
	default:
		fmt.Println("getHandler: args mismatch, please specify [pod|service|job|deploy]/[pods|services|jobs|deploys]")
		fmt.Println("Use like: kubectl get pod [podNamespace]/[podName]")
	}
}

// ==============================================
//
// get pod handler
//
// kubeclt get pod [podNamespace]/[podName]
// 测试命令
// ==============================================
func getPodHandler(cmd *cobra.Command, args []string) {
	if len(args) == 1 {
		// 尝试获取用户是否指定了namespace
		namespace, _ := cmd.Flags().GetString("namespace")

		// 如果没有指定namespace，则使用默认的namespace
		if namespace == "" {
			namespace = config.DefaultNamespace
		}

		// 获取default namespace下的所有Pod
		getNamespacePods(namespace)

	} else if len(args) == 2 {
		// 获取namespace和podName
		namespace, name, err := parseNameAndNamespace(args[1])

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if namespace == "" || name == "" {
			fmt.Println("name of namespace or podName is empty")
			fmt.Println("Use like: kubectl get pod [podNamespace]/[podName]")
			return
		}

		// 获取指定的Pod
		getSpecificPod(namespace, name)

	} else {
		fmt.Println("getHandler: args mismatch, please specify [pod|service|job|deploy]/[pods|services|jobs|deploys]")
		fmt.Println("Use like: kubectl get pod [podNamespace]/[podName]")
	}
}

func getSpecificPod(namespace, name string) {
	url := stringutil.Replace(config.PodSpecURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, name)
	url = config.API_Server_URL_Prefix + url

	pod := &apiObject.PodStore{}
	code, err := netrequest.GetRequestByTarget(url, pod, "data")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if code != http.StatusOK {
		fmt.Println("getSpecificPod: code:", code)
		return
	}

	pods := []apiObject.PodStore{*pod}
	printPodsResult(pods)
}

func getNamespacePods(namespace string) {
	url := stringutil.Replace(config.PodsURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = config.API_Server_URL_Prefix + url

	pods := []apiObject.PodStore{}

	code, err := netrequest.GetRequestByTarget(url, &pods, "data")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if code != http.StatusOK {
		fmt.Println("getNamespacePods: code:", code)
		return
	}

	printPodsResult(pods)
}

// ==============================================
//
// get service handler
//
// kubeclt get service [podNamespace]/[podName]
// 测试命令
// ==============================================

func getServiceHandler(cmd *cobra.Command, args []string) {
	if len(args) == 1 {
		// 尝试获取用户是否指定了namespace
		namespace, _ := cmd.Flags().GetString("namespace")

		// 如果没有指定namespace，则使用默认的namespace
		if namespace == "" {
			namespace = config.DefaultNamespace
		}

		// 获取default namespace下的所有Pod
		getNamespaceServices(namespace)

	} else if len(args) == 2 {
		// 获取namespace和podName
		namespace, name, err := parseNameAndNamespace(args[1])

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if namespace == "" || name == "" {
			fmt.Println("name of namespace or service Name is empty")
			fmt.Println("Use like: kubectl get service [serviceNamespace]/[serviceName]")
			return
		}

		// 获取指定的Pod
		getSpecificService(namespace, name)

	} else {
		fmt.Println("getHandler: args mismatch, please specify [pod|service|job|deploy]/[pods|services|jobs|deploys]")
		fmt.Println("Use like: kubectl get service [serviceNamespace]/[serviceName]")
	}
}

func getSpecificService(namespace, name string) {
	url := stringutil.Replace(config.ServiceSpecURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, name)
	url = config.API_Server_URL_Prefix + url

	service := &apiObject.ServiceStore{}
	code, err := netrequest.GetRequestByTarget(url, service, "data")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if code != http.StatusOK {
		fmt.Println("getSpecificService: code:", code)
		return
	}

	services := []apiObject.ServiceStore{*service}
	printServicesResult(services)

}

func getNamespaceServices(namespace string) {
	url := stringutil.Replace(config.ServiceURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = config.API_Server_URL_Prefix + url

	services := []apiObject.ServiceStore{}

	code, err := netrequest.GetRequestByTarget(url, &services, "data")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if code != http.StatusOK {
		fmt.Println("getNamespaceServices: code:", code)
		return
	}

	printServicesResult(services)
	fmt.Println("")
	printServicesPortInfo(services)
}

func getJobHandler(cmd *cobra.Command, args []string) {
	if len(args) == 1 {
		// 尝试获取用户是否指定了namespace
		namespace, _ := cmd.Flags().GetString("namespace")

		// 如果没有指定namespace，则使用默认的namespace
		if namespace == "" {
			namespace = config.DefaultNamespace
		}

		// 获取default namespace下的所有Job
		getNamespaceJobs(namespace)

	} else if len(args) == 2 {
		// 获取namespace和jobName
		namespace, name, err := parseNameAndNamespace(args[1])

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if namespace == "" || name == "" {
			fmt.Println("name of namespace or job Name is empty")
			fmt.Println("Use like: kubectl get job [jobNamespace]/[jobName]")
			return
		}

		// 获取指定的Job
		getSpecificJob(namespace, name)
	} else {
		fmt.Println("getHandler: args mismatch, please specify [pod|service|job|deploy]/[pods|services|jobs|deploys]")
		fmt.Println("Use like: kubectl get job [jobNamespace]/[jobName]")
	}
}

func getSpecificJob(namespace, name string) {
	url := stringutil.Replace(config.JobSpecURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, name)
	url = config.API_Server_URL_Prefix + url

	job := &apiObject.JobStore{}
	code, err := netrequest.GetRequestByTarget(url, job, "data")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if code != http.StatusOK {
		fmt.Println("getSpecificJob: code:", code)
		return
	}

	jobs := []apiObject.JobStore{*job}
	printJobsResult(jobs)

	if job.Status.State == apiObject.JobState_COMPLETED {
		fileURL := stringutil.Replace(config.JobFileSpecURL, config.URL_PARAM_NAMESPACE_PART, namespace)
		fileURL = stringutil.Replace(fileURL, config.URL_PARAM_NAME_PART, name)
		fileURL = config.API_Server_URL_Prefix + fileURL

		jobFile := &apiObject.JobFile{}
		code, err := netrequest.GetRequestByTarget(fileURL, jobFile, "data")

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if code != http.StatusOK {
			fmt.Println("getSpecificJob: code:", code)
			return
		}

		jobfiles := []apiObject.JobFile{*jobFile}

		printJobOutPutResults(jobfiles)
	}

}

func getNamespaceJobs(namespace string) {
	url := stringutil.Replace(config.JobsURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = config.API_Server_URL_Prefix + url

	fmt.Println(url)

	jobs := []apiObject.JobStore{}

	code, err := netrequest.GetRequestByTarget(url, &jobs, "data")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if code != http.StatusOK {
		fmt.Println("getNamespaceJobs: code:", code)
		return
	}

	printJobsResult(jobs)
}

// ==============================================
//
// get dns handler
//
// kubeclt get dns [DnsNamespace]/[DnsName]
// 测试命令
// ==============================================
func getDnsHandler(cmd *cobra.Command, args []string) {
	if len(args) == 1 {
		// 尝试获取用户是否指定了namespace
		namespace, _ := cmd.Flags().GetString("namespace")

		// 如果没有指定namespace，则使用默认的namespace
		if namespace == "" {
			namespace = config.DefaultNamespace
		}

		// 获取default namespace下的所有Dns
		getNamespaceDns(namespace)

	} else if len(args) == 2 {
		// 获取namespace和podName
		namespace, name, err := parseNameAndNamespace(args[1])

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if namespace == "" && name == "" {
			fmt.Println("name of namespace or dnsName is empty")
			fmt.Println("Use like: kubectl get dns [dnsNamespace]/[dnsName]")
			return
		}

		// 获取指定的Dns
		getSpecificDns(namespace, name)

	} else {
		fmt.Println("getHandler: args mismatch, please specify [pod|service|job|deploy]/[pods|services|jobs|deploys]")
		fmt.Println("Use like: kubectl get dns [podNamespace]/[podName]")
	}
}

func getSpecificDns(namespace, name string) {
	url := stringutil.Replace(config.DnsSpecURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, name)
	url = config.API_Server_URL_Prefix + url

	dns := &apiObject.DnsStore{}
	code, err := netrequest.GetRequestByTarget(url, dns, "data")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if code != http.StatusOK {
		fmt.Println("getSpecificDns: code:", code)
		return
	}
	dnsStores := []apiObject.DnsStore{*dns}
	printDnssResult(dnsStores)
}

func getNamespaceDns(namespace string) {

	url := stringutil.Replace(config.DnsURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = config.API_Server_URL_Prefix + url
	dnsStores := []apiObject.DnsStore{}

	code, err := netrequest.GetRequestByTarget(url, &dnsStores, "data")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if code != http.StatusOK {
		fmt.Println("getNamespaceDnss: code:", code)
		return
	}

	printDnssResult(dnsStores)
}

// ==============================================

// 打印get的结果和报错信息，尽可能对用户友好
// ==============================================
//
//	Colorful Print Functions
//
// ==============================================
// 带有颜色的表格输出

func printPodsResult(pods []apiObject.PodStore) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Kind", "Namespace", "Name", "Status"})

	// 遍历所有的Pod
	for _, pod := range pods {
		printPodResult(&pod, t)
	}

	t.Render()
}

func printPodResult(pod *apiObject.PodStore, t table.Writer) {
	var coloredPodStatus string

	switch pod.Status.Phase {
	case apiObject.PodPending:
		coloredPodStatus = color.YellowString("Pending")
	case apiObject.PodRunning:
		coloredPodStatus = color.GreenString("Running")
	case apiObject.PodSucceeded:
		coloredPodStatus = color.BlueString("Succeeded")
	case apiObject.PodFailed:
		coloredPodStatus = color.RedString("Failed")
	case apiObject.PodUnknown:
		coloredPodStatus = color.YellowString("Unknown")
	default:
		coloredPodStatus = color.YellowString("Unknown")
	}

	// HiCyan
	t.AppendRows([]table.Row{
		{
			color.BlueString(string(Get_Kind_Pod)),
			color.HiCyanString(pod.GetPodNamespace()),
			color.HiCyanString(pod.GetPodName()),
			coloredPodStatus,
		},
	})
}

func printServicesResult(service []apiObject.ServiceStore) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Kind", "Namespace", "Name", "ClusterIP"})

	// 遍历所有的Pod
	for _, s := range service {
		printServiceResult(&s, t)
	}

	t.Render()
}

func printServiceResult(service *apiObject.ServiceStore, t table.Writer) {
	// HiCyan
	t.AppendRows([]table.Row{
		{
			color.BlueString(string(Get_Kind_Service)),
			color.HiCyanString(service.GetNamespace()),
			color.HiCyanString(service.GetName()),
			color.GreenString(service.Spec.ClusterIP),
		},
	})
}

func printServicesPortInfo(service []apiObject.ServiceStore) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Namespace/Name", "ClusterIP", "Port", "EndpointIP/Port", "Protocol"})

	// 遍历所有的Pod
	for _, s := range service {
		printAServicePortInfo(&s, t)
	}

	t.Render()
}

func printAServicePortInfo(service *apiObject.ServiceStore, t table.Writer) {
	// HiCyan
	endpointIPAndPort := ""
	for _, endpoint := range service.Status.Endpoints{
			for _, port := range endpoint.Ports{
				endpointIPAndPort += endpoint.IP + "/" + port + " "
			}
	}
	t.AppendRows([]table.Row{
		{
			color.HiCyanString(service.GetNamespace() + "/" + service.GetName()),
			color.GreenString(service.Spec.ClusterIP),
			color.HiCyanString(strconv.Itoa(int(service.Spec.Ports[0].Port))),
			color.HiCyanString(endpointIPAndPort),
			color.HiCyanString(string(service.Spec.Ports[0].Protocol)),
		},
	})
}

func printJobsResult(jobs []apiObject.JobStore) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Kind", "Namespace", "Name", "Status"})

	// 遍历所有的Pod
	for _, job := range jobs {
		printJobResult(&job, t)
	}

	t.Render()
}

func printJobResult(job *apiObject.JobStore, t table.Writer) {
	var coloredJobStatus string

	switch job.Status.State {
	case apiObject.JobState_BOOT_FAIL:
		coloredJobStatus = color.RedString("BOOT_FAIL")
	case apiObject.JobState_CANCELLED:
		coloredJobStatus = color.YellowString("CANCELLED")
	case apiObject.JobState_COMPLETED:
		coloredJobStatus = color.BlueString("COMPLETED")
	case apiObject.JobState_DEADLINE:
		coloredJobStatus = color.RedString("DEADLINE")
	case apiObject.JobState_FAILED:
		coloredJobStatus = color.RedString("FAILED")
	case apiObject.JobState_NODE_FAIL:
		coloredJobStatus = color.RedString("NODE_FAIL")
	case apiObject.JobState_OUT_OF_MEMORY:
		coloredJobStatus = color.RedString("OUT_OF_MEMORY")
	case apiObject.JobState_PENDING:
		coloredJobStatus = color.YellowString("PENDING")
	case apiObject.JobState_PREEMPTED:
		coloredJobStatus = color.YellowString("PREEMPTED")
	case apiObject.JobState_RUNNING:
		coloredJobStatus = color.GreenString("RUNNING")
	case apiObject.JobState_SUSPENDED:
		coloredJobStatus = color.YellowString("SUSPENDED")
	case apiObject.JobState_TIMEOUT:
		coloredJobStatus = color.RedString("TIMEOUT")
	case apiObject.JobState_COMPLETING:
		coloredJobStatus = color.BlueString("COMPLETING")
	case apiObject.JobState_REVOKED:
		coloredJobStatus = color.RedString("REVOKED")
	default:
		coloredJobStatus = job.Status.State
	}

	// HiCyan
	t.AppendRows([]table.Row{
		{
			color.BlueString(string(Get_Kind_Job)),
			color.HiCyanString(job.GetJobNamespace()),
			color.HiCyanString(job.GetJobName()),
			coloredJobStatus,
		},
	})
}

func printJobOutPutResults(jobfiles []apiObject.JobFile) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"OutType", "Namespace/Name", "Content"})

	// 遍历所有的Pod
	for _, job := range jobfiles {
		printJobOutPutResult(&job, t)
	}

	t.Render()
}

func printJobOutPutResult(jobfile *apiObject.JobFile, t table.Writer) {
	// HiCyan
	t.AppendRows([]table.Row{
		{
			color.GreenString("output"),
			color.HiCyanString(jobfile.GetJobNamespace() + "/" + jobfile.GetJobName()),
			color.GreenString(string(jobfile.OutputFile)),
		},
		{
			color.RedString("error"),
			color.HiCyanString(jobfile.GetJobNamespace() + "/" + jobfile.GetJobName()),
			color.GreenString(string(jobfile.ErrorFile)),
		},
	})
}

func printDnssResult(dnss []apiObject.DnsStore) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Kind", "Namespace", "Name", "Status"})

	// 遍历所有的DnsStore
	for _, dnsStore := range dnss {
		printDnsResult(&dnsStore, t)
	}

	t.Render()
}

func printDnsResult(dns *apiObject.DnsStore, t table.Writer) {
	var coloredDnsStatus string

	switch dns.Status.Phase {
	case apiObject.PodPending:
		coloredDnsStatus = color.YellowString("Pending")
	case apiObject.PodRunning:
		coloredDnsStatus = color.GreenString("Running")
	case apiObject.PodSucceeded:
		coloredDnsStatus = color.BlueString("Succeeded")
	case apiObject.PodFailed:
		coloredDnsStatus = color.RedString("Failed")
	case apiObject.PodUnknown:
		coloredDnsStatus = color.YellowString("Unknown")
	default:
		coloredDnsStatus = color.YellowString("Unknown")
	}

	// HiCyan
	t.AppendRows([]table.Row{
		{
			color.BlueString(string(Get_Kind_Dns)),
			color.HiCyanString(dns.ToDns().GetObjectNamespace()),
			color.HiCyanString(dns.ToDns().GetObjectName()),
			coloredDnsStatus,
		},
	})
}

// args: [podNamespace]/[podName]
// 返回值: podNamespace, podName, error
func parseNameAndNamespace(arg string) (string, string, error) {
	parts := strings.Split(arg, "/")
	if len(parts) != 2 {
		return "", "", errors.New("invalid argument, use like: [podNamespace]/[podName]")
	}

	namespace := parts[0]
	name := parts[1]

	return namespace, name, nil
}
