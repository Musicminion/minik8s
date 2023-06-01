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

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Kubectl update can update apiObject in a declarative way",
	Long:  "Kubectl update can update apiObject in a declarative way, usage kubectl update [file]",
	Run:   updateHandler,
}

type UpdateObject string

// Apply的对象名字
const (
	Update_Kind_Pod        UpdateObject = "Pod"
	Update_Kind_Job        UpdateObject = "Job"
	Update_Kind_Service    UpdateObject = "Service"
	Update_Kind_Replicaset UpdateObject = "Replicaset"
	Update_Kind_Dns        UpdateObject = "Dns"
	Update_Kind_Hpa        UpdateObject = "Hpa"
	Update_Kind_Func       UpdateObject = "Function"
	Update_Kind_Workflow   UpdateObject = "Workflow"
)

func updateHandler(cmd *cobra.Command, args []string) {
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
	case string(Update_Kind_Func):
		updateFuncHandler(fileContent)
	default:
		fmt.Println("not support this kind of object")
	}
}

func updateFuncHandler(fileContent []byte) {
	var function apiObject.Function
	err := kubectlutil.ParseAPIObjectFromYamlfileContent(fileContent, &function)

	if err != nil {
		printUpdateResult(UpdateObject(Update_Kind_Func), ApplyResult_Failed, "parse yaml failed", err.Error())
		return
	}

	if function.Metadata.Name == "" {
		printUpdateResult(UpdateObject(Update_Kind_Func), ApplyResult_Failed, "empty name", "function name is empty")
		return
	}

	if function.Metadata.Namespace == "" {
		function.Metadata.Namespace = config.DefaultNamespace
	}

	fileInfo, err := os.Stat(function.Spec.UserUploadFilePath)
	if !(err == nil && fileInfo.IsDir()) {
		printUpdateResult(UpdateObject(Update_Kind_Func), ApplyResult_Failed, "submit folder not exist", "function submit folder not exist")
		return
	}

	uploadFolder := function.Spec.UserUploadFilePath

	err = zip.CompressToZip(uploadFolder, uploadFolder+".zip")

	if err != nil {
		printUpdateResult(UpdateObject(Update_Kind_Func), ApplyResult_Failed, "zip folder failed", err.Error())
		return
	}

	zipFileBytes, err := os.ReadFile(uploadFolder + ".zip")

	if err != nil {
		printUpdateResult(UpdateObject(Update_Kind_Func), ApplyResult_Failed, "read zip file failed", err.Error())
		return
	}

	function.Spec.UserUploadFile = zipFileBytes

	// 删除产生的zip文件
	err = os.Remove(uploadFolder + ".zip")

	if err != nil {
		printUpdateResult(UpdateObject(Update_Kind_Func), ApplyResult_Failed, "delete zip file failed", err.Error())
		return
	}

	// 发请求
	URL := config.GetAPIServerURLPrefix() + config.FunctionSpecURL
	URL = stringutil.Replace(URL, config.URL_PARAM_NAMESPACE_PART, function.Metadata.Namespace)
	URL = stringutil.Replace(URL, config.URL_PARAM_NAME_PART, function.Metadata.Name)

	code, err, msg := kubectlutil.PutAPIObjectToServer(URL, function)

	if err != nil {
		printUpdateResult(UpdateObject(Update_Kind_Func), ApplyResult_Failed, "post obj failed", err.Error())
		return
	}

	if code == http.StatusOK {
		printUpdateResult(Update_Kind_Func, ApplyResult_Success, "created", msg)
		fmt.Println()
		printUpdateObjectInfo(Update_Kind_Func, function.Metadata.Name, function.Metadata.Namespace)
	} else {
		printUpdateResult(Update_Kind_Func, ApplyResult_Failed, "failed", msg)
	}
}

// ==============================================
// 带有颜色的表格输出
func printUpdateResult(kind UpdateObject, result updateResult, info string, reason string) {
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

func printUpdateObjectInfo(kind UpdateObject, name string, namespace string) {
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
