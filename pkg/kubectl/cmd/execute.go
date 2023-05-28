package cmd

import (
	"encoding/json"
	"fmt"
	"miniK8s/pkg/config"
	netrequest "miniK8s/util/netRequest"

	"github.com/spf13/cobra"
)

var executeCmd = &cobra.Command{
	Use:   "execute",
	Short: "Kubectl execute can execute function in a declarative way",
	Long: "Kubectl execute can execute function in a declarative way, usage kubectl execute [namespace]/[name] [parameters]\n" +
		"For example: kubectl execute default/func1 x=1 y=2",
	Run: executeHandler,
}

func executeHandler(cmd *cobra.Command, args []string) {
	// 【TODO】
	// 读取函数文件
	// 解析函数文件
	// 执行函数
	// 输出函数执行结果
	if len(args) <= 1 {
		fmt.Println("missing some parameters")
		fmt.Println("Use like: kubectl execute" + " [namespace]/[name] [parameters]")
		return
	}

	namespace, name, err := parseNameAndNamespace(args[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	if namespace == "" || name == "" {
		fmt.Println("name of namespace or podName is empty")
		fmt.Println("Use like: kubectl execute" + " [namespace]/[name] [parameters]")
		return
	}
	// 解析参数
	jsonString := args[1]
	var jsonData map[string]interface{}
	err = json.Unmarshal([]byte(jsonString), &jsonData)
	if err != nil {
		fmt.Println("解析JSON出错:", err)
		return
	}

	// 向serveless server发送POST请求
	URL := config.GetServelessServerURLPrefix() + "/" + namespace + "/" + name
	code, res, err := netrequest.PostRequestByTarget(URL, jsonData)
	if err != nil {
		fmt.Println(err)
		return
	}
	if code != 200 {
		fmt.Println("execute function failed, code:", code, "msg: ", res.(map[string]interface{})["msg"])
		return
	}
	fmt.Println("execute function success, result:", res)
}
