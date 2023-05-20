package cmd

import (
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/kubectl/kubectlutil"
	"miniK8s/util/file"
	"miniK8s/util/stringutil"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Kubectl apply can create apiObject in a declarative way",
	Long:  "Kubectl apply can create apiObject in a declarative way, usage kubectl apply -f [file]",
	Run:   applyHandler,
}

type ApplyObject string

// Apply的对象名字
const (
	Apply_Kind_Pod     ApplyObject = "Pod"
	Apply_Kind_Job     ApplyObject = "Job"
	Apply_kind_Service ApplyObject = "Service"
	Apply_kind_Deploy  ApplyObject = "Deployment"
)

// Apply的Result

type ApplyResult string

const (
	ApplyResult_Success ApplyResult = "Success"
	ApplyResult_Failed  ApplyResult = "Failed"
	ApplyResult_Unknow  ApplyResult = "Unknow"
)

func applyHandler(cmd *cobra.Command, args []string) {
	k8log.DebugLog("applyHandler", "args: "+strings.Join(args, " "))
	// 打印出来所有的参数
	// 检查参数的数量是否为1
	if len(args) != 1 {
		cmd.Usage()
		return
	}
	// 检查参数是否是文件 读取文件
	fileInfo, err := os.Stat(args[0])
	if err != nil {
		fmt.Println(err.Error())
		cmd.Usage()
		return
	}
	if fileInfo.IsDir() {
		fmt.Println("file is a directory")
		cmd.Usage()
		return
	}
	// 读取文件的内容
	fileContent, err := file.ReadFile(args[0])
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 解析API对象的种类
	Kind, err := kubectlutil.GetAPIObjectTypeFromYamlFile(fileContent)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	switch Kind {
	case string(Apply_Kind_Pod):
		applyPodHandler(fileContent)
	case string(Apply_kind_Service):
		applyServiceHandler(fileContent)
	case string(Apply_kind_Deploy):
		fmt.Println("Deployment not support now")
	case string(Apply_Kind_Job):
		fmt.Println("Job not support now")
	default:
		fmt.Println("default")
	}
}

// =========================================================
//
// 处理Pod的Apply
//
// =========================================================

func applyPodHandler(fileContent []byte) {
	// fmt.Println("Kind: Pod")
	// 完成YAML转化为POD对象
	var pod apiObject.Pod
	err := kubectlutil.ParseAPIObjectFromYamlfileContent(fileContent, &pod)

	if err != nil {
		printApplyResult(Apply_Kind_Pod, ApplyResult_Failed, "parse yaml failed", err.Error())
		return
	}

	// 检查Pod的名字是否为空
	if pod.GetPodName() == "" {
		printApplyResult(Apply_Kind_Pod, ApplyResult_Failed, "empty name", "pod name is empty")
		return
	}

	// 发请求
	URL := config.API_Server_URL_Prefix + config.PodsURL

	if pod.GetPodNamespace() == "" {
		pod.Metadata.Namespace = config.DefaultNamespace
	}

	URL = stringutil.Replace(URL, config.URL_PARAM_NAMESPACE_PART, pod.GetPodNamespace())

	code, err, msg := kubectlutil.PostAPIObjectToServer(URL, pod)
	if err != nil {
		// fmt.Println(err.Error())
		printApplyResult(Apply_Kind_Pod, ApplyResult_Failed, "post obj failed", err.Error())
		return
	}

	if code == http.StatusCreated {
		printApplyResult(Apply_Kind_Pod, ApplyResult_Success, "created", msg)
		fmt.Println()
		printApplyObjectInfo(Apply_Kind_Pod, pod.GetPodName(), pod.GetPodNamespace())
	} else {
		printApplyResult(Apply_Kind_Pod, ApplyResult_Failed, "failed", msg)
	}
}

// =========================================================
//
// 处理Service的Apply
//
// =========================================================

func applyServiceHandler(fileContent []byte) {
	// fmt.Println("Kind: Service")
	var service apiObject.Service
	err := kubectlutil.ParseAPIObjectFromYamlfileContent(fileContent, &service)

	if err != nil {
		printApplyResult(Apply_kind_Service, ApplyResult_Failed, "parse yaml failed", err.Error())
		return
	}

	// 检查Service的名字是否为空
	if service.Metadata.Name == "" {
		printApplyResult(Apply_kind_Service, ApplyResult_Failed, "empty name", "service name is empty")
		return
	}

	// 检查Service的Namespace是否为空
	if service.Metadata.Namespace == "" {
		service.Metadata.Namespace = config.DefaultNamespace
	}

	// 发请求
	URL := config.API_Server_URL_Prefix + config.ServiceURL
	URL = stringutil.Replace(URL, config.URL_PARAM_NAMESPACE_PART, service.Metadata.Namespace)

	code, err, msg := kubectlutil.PostAPIObjectToServer(URL, service)

	if err != nil {
		printApplyResult(Apply_kind_Service, ApplyResult_Failed, "post obj failed", err.Error())
		return
	}

	if code == http.StatusCreated {
		printApplyResult(Apply_kind_Service, ApplyResult_Success, "created", msg)
		fmt.Println()
		printApplyObjectInfo(Apply_kind_Service, service.Metadata.Name, service.Metadata.Namespace)
	} else {
		printApplyResult(Apply_kind_Service, ApplyResult_Failed, "failed", msg)
	}
}

// 打印Apply的结果和报错信息，尽可能对用户友好
// ==============================================
//
//	Colorful Print Functions
//
// ==============================================
// 带有颜色的表格输出
func printApplyResult(kind ApplyObject, result ApplyResult, info string, reason string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Kind", "Result", "Info", "Reason(Msg)"})

	coloredKind := color.GreenString(string(kind))
	var coloredResult string
	var coloredInfo string
	var coloredReason string

	switch result {
	case ApplyResult_Success:
		coloredResult = color.GreenString(string(result))
	case ApplyResult_Failed:
		coloredResult = color.RedString(string(result))
	case ApplyResult_Unknow:
		coloredResult = color.YellowString(string(result))
	default:
		coloredResult = color.BlueString(string(result))
	}

	coloredInfo = color.CyanString(info)
	coloredReason = color.CyanString(reason)

	t.AppendRows([]table.Row{
		{coloredKind, coloredResult, coloredInfo, coloredReason},
	})

	t.Render()
}

func printApplyObjectInfo(kind ApplyObject, name string, namespace string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Kind", "Name", "Namespace"})

	coloredKind := color.GreenString(string(kind))
	coloredName := color.GreenString(name)
	coloredNamespace := color.GreenString(namespace)

	t.AppendRows([]table.Row{
		{coloredKind, coloredName, coloredNamespace},
	})

	t.Render()
}
