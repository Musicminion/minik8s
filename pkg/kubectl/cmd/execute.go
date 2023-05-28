package cmd

import (
	"fmt"
	"miniK8s/pkg/apiObject"

	"github.com/spf13/cobra"
)

var executeCmd = &cobra.Command{
	Use:   "execute",
	Short: "Kubectl execute can execute function in a declarative way",
	Long:  "Kubectl execute can execute function in a declarative way, usage kubectl execute [function file]",
	Run:   deleteHandler,
}

func executeHandler(cmd *cobra.Command, args []string) {
	// 【TODO】
	// 读取函数文件
	// 解析函数文件
	// 执行函数
	// 输出函数执行结果
	if len(args) <= 1 {
		fmt.Println("getObjectHandler: num of args is wrong, please specify " + apiObject.AllResourceKind)
		cmd.Usage()
		return
	}

	kind := args[0] 

	namespace, name, err := parseNameAndNamespace(args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	if namespace == "" || name == "" {
		fmt.Println("name of namespace or podName is empty")
		fmt.Println("Use like: kubectl get" + kind + "[podNamespace]/[podName]")
		return
	}

	

}
