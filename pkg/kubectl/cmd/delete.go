package cmd

import (
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/kubectl/kubectlutil"
	"miniK8s/util/file"
	"miniK8s/util/stringutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Kubectl delete can delete apiObject in a declarative way",
	Long:  "Kubectl delete can delete apiObject in a declarative way, usage kubectl delete -f [file]",
	Run:   deleteHandler,
}

func deleteHandler(cmd *cobra.Command, args []string) {
	k8log.DebugLog("deleteHandler", "args: "+strings.Join(args, " "))

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
		fmt.Println("Kind: Pod")
		// 完成YAML转化为POD对象
		var pod apiObject.Pod
		kubectlutil.ParseAPIObjectFromYamlfileContent(fileContent, &pod)
		podURL := config.API_Server_URL_Prefix + config.PodSpecURL
		podURL = stringutil.Replace(podURL, config.URL_PARAM_NAMESPACE_PART, pod.GetPodNamespace())
		podURL = stringutil.Replace(podURL, config.URL_PARAM_NAME_PART, pod.GetPodName())

		_, err = kubectlutil.DeleteAPIObjectToServer(podURL)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

	case "Service":
		fmt.Println("Kind: Service")
		var service apiObject.Service
		kubectlutil.ParseAPIObjectFromYamlfileContent(fileContent, &service)
		serviceURL := config.API_Server_URL_Prefix + config.ServiceSpecURL
		serviceURL = stringutil.Replace(serviceURL, config.URL_PARAM_NAMESPACE_PART, service.Metadata.Namespace)
		serviceURL = stringutil.Replace(serviceURL, config.URL_PARAM_NAME_PART, service.Metadata.Name)

		_, err = kubectlutil.DeleteAPIObjectToServer(serviceURL)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

	case "Deployment":
		fmt.Println("Deployment")
	// 其他默认的
	default:
		fmt.Println("default")
	}

	println("delete Handle finish")
}
