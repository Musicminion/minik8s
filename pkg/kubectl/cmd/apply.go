package cmd

import (
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/pkg/kubectl/kubectlutil"
	"miniK8s/util/file"
	"miniK8s/util/stringutil"
	"miniK8s/util/zip"
	"net/http"
	"os"

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
	Apply_Kind_Pod        ApplyObject = "Pod"
	Apply_Kind_Job        ApplyObject = "Job"
	Apply_kind_Service    ApplyObject = "Service"
	Apply_kind_Replicaset ApplyObject = "Replicaset"
	Apply_kind_Dns        ApplyObject = "Dns"
	Apply_kind_Hpa        ApplyObject = "Hpa"
	Apply_kind_Func       ApplyObject = "Function"
	Apply_kind_Workflow   ApplyObject = "Workflow"
)

// Apply的Result

type ApplyResult string

const (
	ApplyResult_Success ApplyResult = "Success"
	ApplyResult_Failed  ApplyResult = "Failed"
	ApplyResult_Unknow  ApplyResult = "Unknow"
)

func applyHandler(cmd *cobra.Command, args []string) {
	// k8log.DebugLog("applyHandler", "args: "+strings.Join(args, " "))
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
	case string(Apply_kind_Replicaset):
		applyRepliacasetHandler(fileContent)
	case string(Apply_Kind_Job):
		applyJobHandler(fileContent)
	case string(Apply_kind_Dns):
		applyDnsHandler(fileContent)
	case string(Apply_kind_Hpa):
		applyHpaHandler(fileContent)
	case string(Apply_kind_Func):
		applyFuncHandler(fileContent)
	case string(Apply_kind_Workflow):
		applyWorkflowHandler(fileContent)
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
	if pod.GetObjectName() == "" {
		printApplyResult(Apply_Kind_Pod, ApplyResult_Failed, "empty name", "pod name is empty")
		return
	}

	// 发请求
	URL := config.GetAPIServerURLPrefix() + config.PodsURL

	if pod.GetObjectNamespace() == "" {
		pod.Metadata.Namespace = config.DefaultNamespace
	}

	URL = stringutil.Replace(URL, config.URL_PARAM_NAMESPACE_PART, pod.GetObjectNamespace())

	code, err, msg := kubectlutil.PostAPIObjectToServer(URL, pod)
	if err != nil {
		// fmt.Println(err.Error())
		printApplyResult(Apply_Kind_Pod, ApplyResult_Failed, "post obj failed", err.Error())
		return
	}

	if code == http.StatusCreated {
		printApplyResult(Apply_Kind_Pod, ApplyResult_Success, "created", msg)
		fmt.Println()
		printApplyObjectInfo(Apply_Kind_Pod, pod.GetObjectName(), pod.GetObjectNamespace())
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
	URL := config.GetAPIServerURLPrefix() + config.ServiceURL
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

// =========================================================
//
// 处理Job的Apply
// 测试用例  go run ./main/ apply ./kubectlutil/testFile/job-with-pwd.yaml
//
// =========================================================

// 逻辑如下
// 1. 检查Job的名字是否为空
// 2. 检查Job的Namespace是否为空
// 3. 上传Job对应yaml信息，创建api对象
// 4. 检查Job对应的文件是否存在
// 5. 上传Job对应的文件
// 6. 检查Job对应的文件是否上传成功

func applyJobHandler(fileContent []byte) {
	// fmt.Println("Kind: Job")
	var job apiObject.Job
	err := kubectlutil.ParseAPIObjectFromYamlfileContent(fileContent, &job)

	if err != nil {
		printApplyResult(Apply_Kind_Job, ApplyResult_Failed, "parse yaml failed", err.Error())
		return
	}

	// 检查Job的名字是否为空
	if job.Metadata.Name == "" {
		printApplyResult(Apply_Kind_Job, ApplyResult_Failed, "empty name", "job name is empty")
		return
	}

	// 检查Job的Namespace是否为空
	if job.Metadata.Namespace == "" {
		job.Metadata.Namespace = config.DefaultNamespace
	}

	// 检查Job对应的文件是否存在
	submitFolder := job.Spec.SubmitDirectory

	if submitFolder == "" {
		printApplyResult(Apply_Kind_Job, ApplyResult_Failed, "empty submit folder", "job submit folder is empty")
		return
	}

	// 检查文件夹是否存在
	// 使用Stat函数检查文件夹是否存在
	fileInfo, err := os.Stat(submitFolder)
	if err == nil && fileInfo.IsDir() {
		// fmt.Println("文件夹存在")
		// 发请求
		URL := config.GetAPIServerURLPrefix() + config.JobsURL
		URL = stringutil.Replace(URL, config.URL_PARAM_NAMESPACE_PART, job.Metadata.Namespace)
		code, err, msg := kubectlutil.PostAPIObjectToServer(URL, job)

		if err != nil {
			printApplyResult(Apply_Kind_Job, ApplyResult_Failed, "post obj failed", err.Error())
			return
		}

		if code != http.StatusCreated {
			printApplyResult(Apply_Kind_Job, ApplyResult_Failed, "failed", msg)
			return
		}

		// 然后将文件夹中的文件压缩为zip文件
		err = zip.CompressToZip(submitFolder, submitFolder+".zip")

		// 如果在这个时候发现错误，就会删除之前的Job
		if err != nil {
			printApplyResult(Apply_Kind_Job, ApplyResult_Failed, "zip folder failed", err.Error())
			delUnusedJob(job.Metadata.Name, job.Metadata.Namespace)
			return
		}

		zipFileBytes, err := os.ReadFile(submitFolder + ".zip")
		// 如果在这个时候发现错误，就会删除之前的Job
		if err != nil {
			printApplyResult(Apply_Kind_Job, ApplyResult_Failed, "read zip file failed", err.Error())
			delUnusedJob(job.Metadata.Name, job.Metadata.Namespace)
			return
		}

		// 然后将zip文件上传到服务器
		userZipFile := apiObject.JobFile{
			Basic: apiObject.Basic{
				Kind: "JobFile",
				Metadata: apiObject.Metadata{
					Name:      job.Metadata.Name,
					Namespace: job.Metadata.Namespace,
				},
			},
			UserUploadFile: zipFileBytes,
		}

		// 然后将userZipFile上传到服务器
		fileURL := config.GetAPIServerURLPrefix() + config.JobFileURL

		code, err, msg = kubectlutil.PostAPIObjectToServer(fileURL, userZipFile)

		if err != nil || code != http.StatusCreated {
			printApplyResult(Apply_Kind_Job, ApplyResult_Failed, "upload zip file failed", err.Error())
			delUnusedJob(job.Metadata.Name, job.Metadata.Namespace)
			return
		}

		// 最后删除zip文件
		_ = os.Remove(submitFolder + ".zip")

		// 打印结果
		printApplyResult(Apply_Kind_Job, ApplyResult_Success, "created", msg)

	} else if os.IsNotExist(err) {
		// fmt.Println("文件夹不存在")
		printApplyResult(Apply_Kind_Job, ApplyResult_Failed, "submit folder not exist", "job submit folder not exist")

	} else if err != nil {
		// fmt.Println("发生错误:", err)
		printApplyResult(Apply_Kind_Job, ApplyResult_Failed, "check submit folder failed", err.Error())
		return
	} else if !fileInfo.IsDir() {
		printApplyResult(Apply_Kind_Job, ApplyResult_Failed, "submit folder not a folder", "job submit folder is not a folder")
		return
	} else {
		printApplyResult(Apply_Kind_Job, ApplyResult_Failed, "unknow error", "unknow error")
		return
	}

}

// =========================================================
//
// 处理Dns的Apply
// 测试用例  go run ./main/ apply ./kubectlutil/testFile/dns.yaml
//
// =========================================================

func applyDnsHandler(fileContent []byte) {
	var dns apiObject.Dns
	err := kubectlutil.ParseAPIObjectFromYamlfileContent(fileContent, &dns)

	if err != nil {
		printApplyResult(Apply_kind_Dns, ApplyResult_Failed, "parse yaml failed", err.Error())
		return
	}

	// 检查Dns的名字是否为空
	if dns.Metadata.Name == "" {
		printApplyResult(Apply_kind_Service, ApplyResult_Failed, "empty name", "dns name is empty")
		return
	}

	// 检查Dns的Namespace是否为空
	if dns.Metadata.Namespace == "" {
		dns.Metadata.Namespace = config.DefaultNamespace
	}

	// 发请求
	URL := config.GetAPIServerURLPrefix() + config.DnsURL
	URL = stringutil.Replace(URL, config.URL_PARAM_NAMESPACE_PART, dns.Metadata.Namespace)

	code, err, msg := kubectlutil.PostAPIObjectToServer(URL, dns)

	if err != nil {
		printApplyResult(Apply_kind_Dns, ApplyResult_Failed, "post obj failed", err.Error())
		return
	}

	if code == http.StatusCreated {
		printApplyResult(Apply_kind_Dns, ApplyResult_Success, "created", msg)
		fmt.Println()
		printApplyObjectInfo(Apply_kind_Dns, dns.Metadata.Name, dns.Metadata.Namespace)
	} else {
		printApplyResult(Apply_kind_Dns, ApplyResult_Failed, "failed", msg)
	}
}

// =========================================================
//
// 处理ReplicaSet的Apply
// 测试用例  go run ./main/ apply ./kubectlutil/testFile/replica.yaml
//
// =========================================================

func applyRepliacasetHandler(fileContent []byte) {
	var repliaset apiObject.ReplicaSet
	err := kubectlutil.ParseAPIObjectFromYamlfileContent(fileContent, &repliaset)

	if err != nil {
		printApplyResult(Apply_kind_Replicaset, ApplyResult_Failed, "parse yaml failed", err.Error())
		return
	}

	// 检查ReplicaSet的名字是否为空
	if repliaset.Metadata.Name == "" {
		printApplyResult(Apply_kind_Replicaset, ApplyResult_Failed, "empty name", "replicaset name is empty")
		return
	}

	// 检查ReplicaSet的Namespace是否为空
	if repliaset.Metadata.Namespace == "" {
		repliaset.Metadata.Namespace = config.DefaultNamespace
	}

	// 发请求

	URL := config.GetAPIServerURLPrefix() + config.ReplicaSetsURL
	URL = stringutil.Replace(URL, config.URL_PARAM_NAMESPACE_PART, repliaset.Metadata.Namespace)

	code, err, msg := kubectlutil.PostAPIObjectToServer(URL, repliaset)

	if err != nil {
		printApplyResult(Apply_kind_Replicaset, ApplyResult_Failed, "post obj failed", err.Error())
		return
	}

	if code == http.StatusCreated {
		printApplyResult(Apply_kind_Replicaset, ApplyResult_Success, "created", msg)
		fmt.Println()
		printApplyObjectInfo(Apply_kind_Replicaset, repliaset.Metadata.Name, repliaset.Metadata.Namespace)
	} else {
		printApplyResult(Apply_kind_Replicaset, ApplyResult_Failed, "failed", msg)
	}
}

// ==============================================
//
// # Apply Func
// go run ./main/ apply func ./
// ==============================================
func applyFuncHandler(fileContent []byte) {
	var function apiObject.Function
	err := kubectlutil.ParseAPIObjectFromYamlfileContent(fileContent, &function)

	if err != nil {
		printApplyResult(Apply_kind_Func, ApplyResult_Failed, "parse yaml failed", err.Error())
		return
	}

	if function.Metadata.Name == "" {
		printApplyResult(Apply_kind_Func, ApplyResult_Failed, "empty name", "function name is empty")
		return
	}

	if function.Metadata.Namespace == "" {
		function.Metadata.Namespace = config.DefaultNamespace
	}

	fileInfo, err := os.Stat(function.Spec.UserUploadFilePath)
	if !(err == nil && fileInfo.IsDir()) {
		printApplyResult(Apply_kind_Func, ApplyResult_Failed, "submit folder not exist", "function submit folder not exist")
		return
	}

	uploadFolder := function.Spec.UserUploadFilePath

	err = zip.CompressToZip(uploadFolder, uploadFolder+".zip")

	if err != nil {
		printApplyResult(Apply_kind_Func, ApplyResult_Failed, "zip folder failed", err.Error())
		return
	}

	zipFileBytes, err := os.ReadFile(uploadFolder + ".zip")

	if err != nil {
		printApplyResult(Apply_kind_Func, ApplyResult_Failed, "read zip file failed", err.Error())
		return
	}

	function.Spec.UserUploadFile = zipFileBytes

	// 删除产生的zip文件
	err = os.Remove(uploadFolder + ".zip")

	if err != nil {
		printApplyResult(Apply_kind_Func, ApplyResult_Failed, "delete zip file failed", err.Error())
		return
	}

	// 发请求
	URL := config.GetAPIServerURLPrefix() + config.FunctionURL
	URL = stringutil.Replace(URL, config.URL_PARAM_NAMESPACE_PART, function.Metadata.Namespace)

	code, err, msg := kubectlutil.PostAPIObjectToServer(URL, function)

	if err != nil {
		printApplyResult(Apply_kind_Func, ApplyResult_Failed, "post obj failed", err.Error())
		return
	}

	if code == http.StatusCreated {
		printApplyResult(Apply_kind_Func, ApplyResult_Success, "created", msg)
		fmt.Println()
		printApplyObjectInfo(Apply_kind_Func, function.Metadata.Name, function.Metadata.Namespace)
	} else {
		printApplyResult(Apply_kind_Func, ApplyResult_Failed, "failed", msg)
	}
}

// =========================================================
//
// 处理Hpa的Apply
// 测试用例  go run ./main/ apply ./kubectlutil/testFile/hpa.yaml
//
// =========================================================

func applyHpaHandler(fileContent []byte) {
	var hpa apiObject.HPA
	err := kubectlutil.ParseAPIObjectFromYamlfileContent(fileContent, &hpa)

	if err != nil {
		printApplyResult(Apply_kind_Hpa, ApplyResult_Failed, "parse yaml failed", err.Error())
		return
	}

	// 检查Dns的名字是否为空
	if hpa.Metadata.Name == "" {
		printApplyResult(Apply_kind_Service, ApplyResult_Failed, "empty name", "hpa name is empty")
		return
	}

	// 检查Dns的Namespace是否为空
	if hpa.Metadata.Namespace == "" {
		hpa.Metadata.Namespace = config.DefaultNamespace
	}

	// 发请求
	URL := config.GetAPIServerURLPrefix() + config.HPAURL
	URL = stringutil.Replace(URL, config.URL_PARAM_NAMESPACE_PART, hpa.Metadata.Namespace)

	code, err, msg := kubectlutil.PostAPIObjectToServer(URL, hpa)

	if err != nil {
		printApplyResult(Apply_kind_Hpa, ApplyResult_Failed, "post obj failed", err.Error())
		return
	}

	if code == http.StatusCreated {
		printApplyResult(Apply_kind_Dns, ApplyResult_Success, "created", msg)
		fmt.Println()
		printApplyObjectInfo(Apply_kind_Hpa, hpa.Metadata.Name, hpa.Metadata.Namespace)
	} else {
		printApplyResult(Apply_kind_Hpa, ApplyResult_Failed, "failed", msg)
	}
}

// ==============================================
//
// 处理Workflow的Apply
//
//
// ==============================================

func applyWorkflowHandler(fileContent []byte) {
	var workflow apiObject.Workflow
	err := kubectlutil.ParseAPIObjectFromYamlfileContent(fileContent, &workflow)

	if err != nil {
		printApplyResult(Apply_kind_Workflow, ApplyResult_Failed, "parse yaml failed", err.Error())
		return
	}

	// 检查Workflow的名字是否为空
	if workflow.Metadata.Name == "" {
		printApplyResult(Apply_kind_Workflow, ApplyResult_Failed, "empty name", "workflow name is empty")
		return
	}

	URL := config.GetAPIServerURLPrefix() + config.WorkflowURL

	if workflow.Metadata.Namespace == "" {
		workflow.Metadata.Namespace = config.DefaultNamespace
	}

	URL = stringutil.Replace(URL, config.URL_PARAM_NAMESPACE_PART, workflow.Metadata.Namespace)

	code, err, msg := kubectlutil.PostAPIObjectToServer(URL, workflow)

	if err != nil {
		printApplyResult(Apply_kind_Workflow, ApplyResult_Failed, "post obj failed", err.Error())
		return
	}

	if code == http.StatusCreated {
		printApplyResult(Apply_kind_Workflow, ApplyResult_Success, "created", msg)
		fmt.Println()
		printApplyObjectInfo(Apply_kind_Workflow, workflow.Metadata.Name, workflow.Metadata.Namespace)
	} else {
		printApplyResult(Apply_kind_Workflow, ApplyResult_Failed, "failed", msg)
	}

}

// ==============================================

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

// ==============================================
//
//	Other Util Functions
//
// ==============================================
func delUnusedJob(name string, namespace string) {
	URL := config.GetAPIServerURLPrefix() + config.JobSpecURL
	URL = stringutil.Replace(URL, config.URL_PARAM_NAMESPACE_PART, namespace)
	URL = stringutil.Replace(URL, config.URL_PARAM_NAME_PART, name)

	code, err := kubectlutil.DeleteAPIObjectToServer(URL)

	if err != nil {
		// fmt.Println("Delete unused job failed: ", err)
		return
	}

	if code != http.StatusNoContent {
		// fmt.Println("Delete unused job failed: ")
		return
	}
}
