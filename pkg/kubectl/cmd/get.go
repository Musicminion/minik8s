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
	"time"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Kubectl get can get apiObject in a declarative way",
	Long:  "Kubectl get can get apiObject in a declarative way, usage kubectl get " + apiObject.AllResourceKind,
	Run:   getObjectHandler,
}

var getNamespaceObjectFuncMap = make(map[string]func(namespace string))
var getSpecificObjectFunMap = make(map[string]func(namespace string, name string))
var getNoNamespaceObjectFuncMap = make(map[string]func())

func init() {
	getCmd.PersistentFlags().StringP("namespace", "n", "", "Namespace")

	// 构建kind到函数的映射
	getNamespaceObjectFuncMap[string(Get_Kind_Pod)] = getNamespacePods
	getNamespaceObjectFuncMap[string(Get_Kind_Service)] = getNamespaceServices
	getNamespaceObjectFuncMap[string(Get_Kind_Job)] = getNamespaceJobs
	getNamespaceObjectFuncMap[string(Get_Kind_Replicaset)] = getNamespaceReplicaSets
	getNamespaceObjectFuncMap[string(Get_Kind_Hpa)] = getNamespaceHpas
	getNamespaceObjectFuncMap[string(Get_Kind_Function)] = getNamespaceFunctions
	getNamespaceObjectFuncMap[string(Get_Kind_Dns)] = getNamespaceDns
	getNamespaceObjectFuncMap[string(Get_Kind_Workflow)] = getNamespaceWorkflows
	
	
	getSpecificObjectFunMap[string(Get_Kind_Pod)] = getSpecificPod
	getSpecificObjectFunMap[string(Get_Kind_Service)] = getSpecificService
	getSpecificObjectFunMap[string(Get_Kind_Job)] = getSpecificJob
	getSpecificObjectFunMap[string(Get_Kind_Replicaset)] = getSpecificReplicaSet
	getSpecificObjectFunMap[string(Get_Kind_Hpa)] = getSpecificHpa
	getSpecificObjectFunMap[string(Get_Kind_Function)] = getSpecificFunction
	getSpecificObjectFunMap[string(Get_Kind_Dns)] = getSpecificDns
	getSpecificObjectFunMap[string(Get_Kind_Workflow)] = getSpecificWorkflow
	
	getNoNamespaceObjectFuncMap[string(Get_Kind_Node)] = getNodes
}

type GetObject string

const (
	Get_Kind_Node       GetObject = "node"
	Get_Kind_Pod        GetObject = "pod"
	Get_Kind_Service    GetObject = "service"
	Get_Kind_Job        GetObject = "job"
	Get_Kind_Replicaset GetObject = "replicaset"
	Get_Kind_Dns        GetObject = "dns"
	Get_Kind_Hpa        GetObject = "hpa"
	Get_Kind_Function   GetObject = "function"
	Get_Kind_Workflow   GetObject = "workflow"
)

func getObjectHandler(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("getObjectHandler: no args, please specify " + apiObject.AllResourceKind)
		cmd.Usage()
		return
	}
	kind := args[0]
	// 判断kind是否在apiObject.AllResourceKind中
	if stringutil.ContainsString(apiObject.AllResourceKindSlice, kind) {
		fmt.Println("getObjectHandler: args mismatch, please specify " + apiObject.AllResourceKind)
		fmt.Println("Use like: kubectl get pod [podNamespace]/[podName]")
		return
	}

	if len(args) == 1 {
		// 如果获取的资源是node，则直接获取node
		if kind == string(Get_Kind_Node) {
			getNodes()
			return 
		}

		// 尝试获取用户是否指定了namespace
		namespace, _ := cmd.Flags().GetString("namespace")

		// 如果没有指定namespace，则使用默认的namespace
		if namespace == "" {
			namespace = config.DefaultNamespace
		}
		
		// 获取default namespace下的所有指定kind的对象
		getNamespaceObjectFuncMap[kind](namespace)

	} else if len(args) == 2 {
		// 获取namespace和podName
		namespace, name, err := parseNameAndNamespace(args[1])

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if namespace == "" || name == "" {
			fmt.Println("name of namespace or podName is empty")
			fmt.Println("Use like: kubectl get" + kind + "[podNamespace]/[podName]")
			return
		}

		// 获取指定的Pod
		getSpecificObjectFunMap[kind](namespace, name)

	} else {
		fmt.Println("getHandler: args mismatch, please specify " + apiObject.AllResourceKind)
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

func getSpecificPod(namespace, name string) {
	url := stringutil.Replace(config.PodSpecURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, name)
	url = config.GetAPIServerURLPrefix() + url

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
	url = config.GetAPIServerURLPrefix() + url

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

func getNodes(){
	url := config.GetAPIServerURLPrefix() + config.NodesURL

	nodes := []apiObject.NodeStore{}

	code, err := netrequest.GetRequestByTarget(url, &nodes, "data")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if code != http.StatusOK {
		fmt.Println("getNodes: code:", code)
		return
	}

	printNodesResult(nodes)
}

// ==============================================
//
// get service handler
//
// kubeclt get service [podNamespace]/[podName]
// 测试命令
// ==============================================

func getSpecificService(namespace, name string) {
	url := stringutil.Replace(config.ServiceSpecURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, name)
	url = config.GetAPIServerURLPrefix() + url

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
	url = config.GetAPIServerURLPrefix() + url

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

func getSpecificJob(namespace, name string) {
	url := stringutil.Replace(config.JobSpecURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, name)
	url = config.GetAPIServerURLPrefix() + url

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
		fileURL = config.GetAPIServerURLPrefix() + fileURL

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
	url = config.GetAPIServerURLPrefix() + url

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

func getSpecificDns(namespace, name string) {
	url := stringutil.Replace(config.DnsSpecURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, name)
	url = config.GetAPIServerURLPrefix() + url

	dns := &apiObject.HpaStore{}
	code, err := netrequest.GetRequestByTarget(url, dns, "data")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if code != http.StatusOK {
		fmt.Println("getSpecificDns: code:", code)
		return
	}
	dnsStores := []apiObject.HpaStore{*dns}
	printDnssResult(dnsStores)
}

func getNamespaceDns(namespace string) {

	url := stringutil.Replace(config.DnsURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = config.GetAPIServerURLPrefix() + url
	dnsStores := []apiObject.HpaStore{}

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
//
// get hpa handler
//
// kubeclt get hpa [HpaNamespace]/[HpaName]
// 测试命令
// ==============================================

func getSpecificHpa(namespace, name string) {
	url := stringutil.Replace(config.HPASpecURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, name)
	url = config.GetAPIServerURLPrefix() + url

	hpa := &apiObject.HPAStore{}
	code, err := netrequest.GetRequestByTarget(url, hpa, "data")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if code != http.StatusOK {
		fmt.Println("getSpecificHpa: code:", code)
		return
	}
	hpaStores := []apiObject.HPAStore{*hpa}
	printHpasResult(hpaStores)
}

func getNamespaceHpas(namespace string) {

	url := stringutil.Replace(config.HPAURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = config.GetAPIServerURLPrefix() + url
	hpaStores := []apiObject.HPAStore{}

	code, err := netrequest.GetRequestByTarget(url, &hpaStores, "data")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if code != http.StatusOK {
		fmt.Println("getNamespaceHpas: code:", code)
		return
	}

	printHpasResult(hpaStores)
}

// ==============================================
//
// get replicaset handler
//
// kubeclt get replicaset [Namespace]/[Name]
// 测试命令
// ==============================================

func getSpecificReplicaSet(namespace, name string) {
	url := stringutil.Replace(config.ReplicaSetSpecURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, name)
	url = config.GetAPIServerURLPrefix() + url

	replicaset := &apiObject.ReplicaSetStore{}
	code, err := netrequest.GetRequestByTarget(url, replicaset, "data")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if code != http.StatusOK {
		fmt.Println("getSpecificDns: code:", code)
		return
	}
	replicasetStores := []apiObject.ReplicaSetStore{*replicaset}
	printReplicasetsResult(replicasetStores)
}

func getNamespaceReplicaSets(namespace string) {

	url := stringutil.Replace(config.ReplicaSetsURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = config.GetAPIServerURLPrefix() + url
	replicasetStores := []apiObject.ReplicaSetStore{}

	code, err := netrequest.GetRequestByTarget(url, &replicasetStores, "data")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if code != http.StatusOK && code != http.StatusNoContent {
		fmt.Println("getNamespaceReplicaSets: code:", code)
		return
	}

	printReplicasetsResult(replicasetStores)
}

// ==============================================
//
// get function handler
//
// kubeclt get hpa [HpaNamespace]/[HpaName]
// 测试命令
// ==============================================

func getSpecificFunction(namespace, name string) {
	url := stringutil.Replace(config.FunctionSpecURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, name)
	url = config.GetAPIServerURLPrefix() + url

	function := &apiObject.Function{}
	code, err := netrequest.GetRequestByTarget(url, function, "data")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if code != http.StatusOK {
		fmt.Println("getSpecificFunction: code:", code)
		return
	}
	functions := []apiObject.Function{*function}
	printFunctionsResult(functions)
}

func getNamespaceFunctions(namespace string) {

	url := stringutil.Replace(config.FunctionURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = config.GetAPIServerURLPrefix() + url
	functions := []apiObject.Function{}

	code, err := netrequest.GetRequestByTarget(url, &functions, "data")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if code != http.StatusOK {
		fmt.Println("getNamespaceFunctions: code:", code)
		return
	}

	printFunctionsResult(functions)
}

// ==============================================
// get Workflow handler
// ==============================================

func getSpecificWorkflow(namespace, name string) {
	url := stringutil.Replace(config.WorkflowSpecURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, name)
	url = config.GetAPIServerURLPrefix() + url

	workflow := &apiObject.WorkflowStore{}
	code, err := netrequest.GetRequestByTarget(url, workflow, "data")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if code != http.StatusOK {
		fmt.Println("getSpecificService: code:", code)
		return
	}

	workflows := []apiObject.WorkflowStore{*workflow}
	printWorkflowsResult(workflows)

}

func getNamespaceWorkflows(namespace string) {
	url := stringutil.Replace(config.WorkflowURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = config.GetAPIServerURLPrefix() + url

	workflows := []apiObject.WorkflowStore{}

	code, err := netrequest.GetRequestByTarget(url, &workflows, "data")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if code != http.StatusOK {
		fmt.Println("getNamespaceWorkflows: code:", code)
		return
	}

	printWorkflowsResult(workflows)
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
	t.AppendHeader(table.Row{"Kind", "Namespace", "Name", "Status", "IP", "RunTime", "Node"})

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

	// 把string转换为time.Time类型
	var createdTime time.Time
	var currentTime time.Time
	var runTime string
	if len(pod.Status.ContainerStatuses) != 0 {
		createdTime, _ = time.Parse(time.RFC3339, pod.Status.ContainerStatuses[0].StartedAt)
		currentTime = time.Now()
		// 得到运行时间，格式： 小时:分钟:秒
		runTime = currentTime.Sub(createdTime).Truncate(time.Second).String()
	} else {
		runTime = "Not Created Yet"
	} // HiCyan
	t.AppendRows([]table.Row{
		{
			color.BlueString(string(Get_Kind_Pod)),
			color.HiCyanString(pod.GetPodNamespace()),
			color.HiCyanString(pod.GetPodName()),
			coloredPodStatus,
			color.GreenString(pod.Status.PodIP),
			color.HiCyanString(runTime),
			color.HiCyanString(pod.Spec.NodeName),
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
	for _, endpoint := range service.Status.Endpoints {
		for _, port := range endpoint.Ports {
			if port == strconv.Itoa(service.Spec.Ports[0].TargetPort) {
				endpointIPAndPort += endpoint.IP + "/" + port + " "
			}
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

func printDnssResult(dnss []apiObject.HpaStore) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Kind", "Namespace", "Name", "Status"})

	// 遍历所有的DnsStore
	for _, dnsStore := range dnss {
		printDnsResult(&dnsStore, t)
	}

	t.Render()
}

func printDnsResult(dns *apiObject.HpaStore, t table.Writer) {
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

func printHpasResult(hpas []apiObject.HPAStore) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Kind", "Namespace", "Name", "CurMem/TargetMem", "CurCpu/TargetCpu"})

	// 遍历所有的DnsStore
	for _, hpaStore := range hpas {
		printHpaResult(&hpaStore, t)
	}

	t.Render()
}

func printHpaResult(hpa *apiObject.HPAStore, t table.Writer) {
	var curCPUPercent = hpa.Status.CurCPUPercent
	var targetCPUPercent = hpa.Spec.Metrics.CPUPercent
	var curMemPercent = hpa.Status.CurMemPercent
	var targetMemPercent = hpa.Spec.Metrics.MemPercent
	// HiCyan
	t.AppendRows([]table.Row{
		{
			color.BlueString(string(Get_Kind_Dns)),
			color.HiCyanString(hpa.ToHPA().GetObjectNamespace()),
			color.HiCyanString(hpa.ToHPA().GetObjectName()),
			color.GreenString(fmt.Sprintf("%.1f%%/%.1f%%", curMemPercent*100, targetMemPercent*100)),
			color.GreenString(fmt.Sprintf("%.1f%%/%.1f%%", curCPUPercent*100, targetCPUPercent*100)),
		},
	})
}

func printReplicasetsResult(replicasets []apiObject.ReplicaSetStore) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Kind", "Namespace", "Name", "Cur/Expect replica"})

	// 遍历所有的DnsStore
	for _, replicaset := range replicasets {
		printReplicasetResult(&replicaset, t)
	}

	t.Render()
}

func printReplicasetResult(replicaset *apiObject.ReplicaSetStore, t table.Writer) {
	// HiCyan
	t.AppendRows([]table.Row{
		{
			color.BlueString(string(Get_Kind_Replicaset)),
			color.HiCyanString(replicaset.ToReplicaSet().GetObjectNamespace()),
			color.HiCyanString(replicaset.ToReplicaSet().GetObjectName()),
			color.GreenString(fmt.Sprintf("\t\t%d/%d", replicaset.Status.ReadyReplicas, replicaset.Status.Replicas)),
		},
	})
}

func printFunctionsResult(functions []apiObject.Function) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Kind", "Namespace", "Name", "FilePath"})

	// 遍历所有的DnsStore
	for _, function := range functions {
		printFunctionResult(&function, t)
	}

	t.Render()
}

func printFunctionResult(function *apiObject.Function, t table.Writer) {
	// HiCyan
	t.AppendRows([]table.Row{
		{
			color.BlueString(string(Get_Kind_Function)),
			color.HiCyanString(function.GetObjectNamespace()),
			color.HiCyanString(function.GetObjectName()),
			color.GreenString(function.Spec.UserUploadFilePath),
		},
	})
}

func printWorkflowsResult(workflows []apiObject.WorkflowStore) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Kind", "Namespace", "Name", "Phase", "Result"})

	// 遍历所有的DnsStore
	for _, workflow := range workflows {
		printWorkflowResult(&workflow, t)
	}

	t.Render()
}

func printWorkflowResult(workflow *apiObject.WorkflowStore, t table.Writer) {
	// HiCyan
	t.AppendRows([]table.Row{
		{
			color.BlueString(string(Get_Kind_Function)),
			color.HiCyanString(workflow.ToWorkflow().ToWorkflowStore().GetNamespace()),
			color.HiCyanString(workflow.ToWorkflow().GetObjectName()),
			color.GreenString(workflow.Status.Phase),
			color.GreenString(workflow.Status.Result),
		},
	})
}

func printNodesResult(nodes []apiObject.NodeStore) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Kind", "Name", "Status", "IP", "CPU", "Memory"})

	// 遍历所有的DnsStore
	for _, node := range nodes {
		printNodeResult(&node, t)
	}

	t.Render()
}

func printNodeResult(node *apiObject.NodeStore, t table.Writer) {
	// HiCyan
	t.AppendRows([]table.Row{
		{
			color.BlueString(string(Get_Kind_Node)),
			color.HiCyanString(node.ToNode().GetObjectName()),
			color.GreenString(string(node.Status.Condition)),
			color.GreenString(node.Status.Ip),
			color.GreenString(strconv.FormatFloat(node.Status.CpuPercent, 'f', 1, 64) + "%"),
			color.GreenString(strconv.FormatFloat(node.Status.MemPercent, 'f', 1, 64) + "%"),
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
