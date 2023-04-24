package cmd

import (
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/pkg/kubectl/kubectlutil"
	"miniK8s/util/file"
	"os"

	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Kubectl apply can create apiObject in a declarative way",
	Long:  "Kubectl apply can create apiObject in a declarative way, usage kubectl apply -f [file]",
	Run:   applyHandler,
}

func applyHandler(cmd *cobra.Command, args []string) {
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
	case "Pod":
		fmt.Println("Pod")
		// 完成YAML转化为POD对象
		var pod apiObject.Pod
		kubectlutil.ParseAPIObjectFromYamlfileContent(fileContent, &pod)
		// // 发请求，走你！
		URL := config.PodsURL
		kubectlutil.PostAPIObjectToServer(URL, pod)
	case "Service":
		fmt.Println("Service")
		// 后面的类似写就完事了！！

	case "Deployment":
		fmt.Println("Deployment")
	// 其他默认的
	default:
		fmt.Println("default")
	}

	println("apply Handle finish")
}
