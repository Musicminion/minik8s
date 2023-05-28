package cmd

import (
	"encoding/json"
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	netrequest "miniK8s/util/netRequest"
	"miniK8s/util/stringutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Show details of a specific resource or group of resources",
	Long:  `Show details of a specific resource or group of resources, usage kubectl describe <resource> <name>`,
	Run:   describeKindHandler,
}

func init() {
	describeCmd.PersistentFlags().StringP("namespace", "n", "", "Namespace")
}

type DescribeObject string

// ==============================================
// 通过kind describe资源
// ==============================================

func describeKindHandler(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("describeHandler: no args, please specify [pod|service|job|replicaset]/[pods|services|jobs|replicasets]")
		cmd.Usage()
		return
	}

	kind := strings.ToLower(args[0])
	// kind = strings.Title(kind)
	tag := language.English
	kind = cases.Title(tag).String(kind)

	if len(args) == 1 {
		// 获取namespace
		namespace := config.DefaultNamespace
		describeNamespaceObjects(kind, namespace)
	} else if len(args) == 2 {
		// 获取namespace和podName
		namespace, name, err := parseNameAndNamespace(args[1])
		if err != nil {
			fmt.Println("describePodHandler: parseNameAndNamespace error:", err)
			return
		}
		if namespace == "" || name == "" {
			fmt.Println("describePodHandler: namespace or name is empty")
			fmt.Println("Use like: kubectl describe pod [podNamespace]/[podName]")
			return
		}
		// 获取特定的pod
		describeSpecificObject(kind, namespace, name)
	} else {
		fmt.Println("describePodHandler: args mismatch, please specify [podNamespace]/[podName]")
		fmt.Println("Use like: kubectl describe pod [podNamespace]/[podName]")
	}
}

func describeNamespaceObjects(kind, namespace string) error {
	url := config.GetAPIServerURLPrefix() + config.ApiResourceMap[kind]
	url = stringutil.Replace(url, config.URL_PARAM_NAMESPACE_PART, namespace)

	// 根据 Kind 类型从映射中查找相应的结构体类型
	structType, ok := apiObject.KindToStructType[kind]
	if !ok {
		return errors.Errorf("Unsupported Kind: %s", kind)
	}

	// 创建对应类型的切片
	objs := reflect.New(reflect.SliceOf(structType)).Interface()

	code, err := netrequest.GetRequestByTarget(url, &objs, "data")

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	if code != http.StatusOK {
		fmt.Println("getNamespacePods: code:", code)
		return err
	}

	// 格式化输出JSON到控制台
	indentedJSON, err := json.MarshalIndent(objs, "", "  ")
	if err != nil {
		fmt.Println("JSON marshaling failed:", err)
		return err
	}
	fmt.Println(string(indentedJSON))
	return nil
}

func describeSpecificObject(kind, namespace, name string) error {
	url := config.GetAPIServerURLPrefix() + config.ApiSpecResourceMap[kind]
	url = stringutil.Replace(url, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, name)
	fmt.Printf("describeSpecificObject: url: %s\n", url)

	structType, ok := apiObject.KindToStructType[kind]
	if !ok {
		return errors.Errorf("Unsupported Kind: %s", kind)
	}

	// 根据结构体类型创建对应的空结构体
	obj := reflect.New(structType).Interface().(apiObject.APIObject)

	code, err := netrequest.GetRequestByTarget(url, &obj, "data")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	if code != http.StatusOK {
		fmt.Println("getNamespacePods: code:", code)
		return err
	}

	// 格式化输出JSON到控制台
	indentedJSON, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		fmt.Println("JSON marshaling failed:", err)
		return err
	}
	fmt.Println(string(indentedJSON))
	return nil
}
