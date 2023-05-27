package cmd

import (
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/pkg/kubectl/kubectlutil"
	"miniK8s/util/file"
	"miniK8s/util/stringutil"
	"net/http"
	"os"
	"reflect"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/table"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Kubectl delete can delete apiObject in a declarative way",
	Long:  "Kubectl delete can delete apiObject in a declarative way, usage kubectl delete -f [file]",
	Run:   deleteHandler,
}

type DeleteObject string
type DeleteResult string

const (
	DeleteResult_Success DeleteResult = "Success"
	DeleteResult_Failed  DeleteResult = "Failed"
	DeleteResult_Unknow  DeleteResult = "Unknow"
)

func DeleteAPIObjectByKind(kind string, yamlContent []byte) error {
	// 根据 Kind 类型从映射中查找相应的结构体类型
	structType, ok := apiObject.KindToStructType[kind]
	if !ok {
		return errors.Errorf("Unsupported Kind: %s", kind)
	}

	// 根据结构体类型创建对应的空结构体
	obj := reflect.New(structType).Interface().(apiObject.APIObject)

	// 解析 YAML 内容到空结构体中
	err := kubectlutil.ParseAPIObjectFromYamlfileContent(yamlContent, obj)
	if err != nil {
		return errors.Wrapf(err, "Failed to parse %s YAML file content", kind)
	}

	// 构造删除 API 对象的 URL
	namespace := obj.GetObjectNamespace()
	if namespace == "" {
		namespace = config.DefaultNamespace
	}
	name := obj.GetObjectName()
	if name == "" {
		return errors.Errorf("Failed to get %s name", kind)
	}

	url := config.GetAPIServerURLPrefix() + config.ApiSpecResourceMap[kind]
	url = stringutil.Replace(url, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, name)

	// 向服务器发送删除请求
	code, err := kubectlutil.DeleteAPIObjectToServer(url)
	if err != nil {
		return errors.Wrapf(err, "Failed to delete %s %s", kind, obj.GetObjectName())
	}
	if code != http.StatusNoContent {
		return errors.Errorf("Failed to delete %s %s, code: %d", kind, obj.GetObjectName(), code)
	}

	return nil
}

func deleteHandler(cmd *cobra.Command, args []string) {
	// k8log.DebugLog("deleteHandler", "args: "+strings.Join(args, " "))

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

	// 根据API对象的种类，删除API对象
	err = DeleteAPIObjectByKind(Kind, fileContent)
	if err != nil {
		printDeleteResult(DeleteObject(Kind), DeleteResult_Failed, "delete obj failed", err.Error())
		return
	}

	printDeleteResult(DeleteObject(Kind), DeleteResult_Success, "delete obj success", "")

}

// ==============================================

// 打印Delete的结果和报错信息，尽可能对用户友好
// ==============================================
//
//	Colorful Print Functions
//
// ==============================================
// 带有颜色的表格输出
func printDeleteResult(kind DeleteObject, result DeleteResult, info string, reason string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Kind", "Result", "Info", "Reason(Msg)"})

	coloredKind := color.GreenString(string(kind))
	var coloredResult string
	var coloredInfo string
	var coloredReason string

	switch result {
	case DeleteResult_Success:
		coloredResult = color.GreenString(string(result))
	case DeleteResult_Failed:
		coloredResult = color.RedString(string(result))
	case DeleteResult_Unknow:
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
