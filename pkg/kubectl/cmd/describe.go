package cmd

import (
	"encoding/json"
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	netrequest "miniK8s/util/netRequest"
	"miniK8s/util/stringutil"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Show details of a specific resource or group of resources",
	Long:  `Show details of a specific resource or group of resources, usage kubectl describe <resource> <name>`,
	Run:   describeHandler,
}

func init() {
	describeCmd.PersistentFlags().StringP("namespace", "n", "", "Namespace")
}

type DescribeObject string

const (
	Describe_Kind_Pod        = "pod"
	Describe_Kind_Service    = "service"
	Describe_Kind_Job        = "job"
	Describe_Kind_Replicaset = "replicaset"

	Describe_Kind_Pods        = "pods"
	Describe_Kind_Services    = "services"
	Describe_Kind_Jobs        = "jobs"
	Describe_Kind_Replicasets = "replicasets"
)

func describeHandler(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("describeHandler: no args, please specify [pod|service|job|replicaset]/[pods|services|jobs|replicasets]")
		cmd.Usage()
		return
	}

	args[0] = strings.ToLower(args[0])

	switch args[0] {
	case string(Describe_Kind_Pod), string(Describe_Kind_Pods):
		describePodHandler(cmd, args)
	case string(Describe_Kind_Service), string(Describe_Kind_Services):
		describeServiceHandler(cmd, args)
	case string(Describe_Kind_Job), string(Describe_Kind_Jobs):
		describeJobHandler(cmd, args)
	case string(Describe_Kind_Replicaset), string(Describe_Kind_Replicasets):
		describeReplicasetHandler(cmd, args)
	default:
		fmt.Println("describeHandler: args mismatch, please specify [pod|service|job|deploy]/[pods|services|jobs|deploys]")
		fmt.Println("Use like: kubectl describe pod <podNamespace>/<pod-name>")
	}

}

// ==============================================
//
// describe pod handler
//
// kubeclt describe pod [podNamespace]/[podName]
// 测试命令
// ==============================================
func describePodHandler(cmd *cobra.Command, args []string) {
	if len(args) == 1 {
		// 尝试获取用户是否制定了namespace
		namespace, _ := cmd.Flags().GetString("namespace")

		// 获取pod的namespace和name
		if namespace == "" {
			namespace = config.DefaultNamespace
		}

		// 获取default namespace下的所有pod
		describeNamespacePods(namespace)

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
		describeSpecificPod(namespace, name)
	} else {
		fmt.Println("describePodHandler: args mismatch, please specify [podNamespace]/[podName]")
		fmt.Println("Use like: kubectl describe pod [podNamespace]/[podName]")
	}

}

func describeNamespacePods(namespace string) {
	url := stringutil.Replace(config.PodsURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = config.API_Server_URL_Prefix + url

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

	// 格式化输出JSON到控制台
	indentedJSON, err := json.MarshalIndent(pods, "", "  ")
	if err != nil {
		fmt.Println("JSON marshaling failed:", err)
		return
	}
	fmt.Println(string(indentedJSON))
}

func describeSpecificPod(namespace, name string) {
	url := stringutil.Replace(config.PodSpecURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, name)
	url = config.API_Server_URL_Prefix + url

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

	// 格式化输出JSON到控制台
	indentedJSON, err := json.MarshalIndent(pod, "", "  ")
	if err != nil {
		fmt.Println("JSON marshaling failed:", err)
		return
	}

	fmt.Println(string(indentedJSON))
}

// ==============================================
//
// describe service handler
//
// ==============================================
func describeServiceHandler(cmd *cobra.Command, args []string) {
	if len(args) == 1 {
		// 尝试获取用户是否制定了namespace
		namespace, _ := cmd.Flags().GetString("namespace")

		if namespace == "" {
			namespace = config.DefaultNamespace
		}

		// 获取default namespace下的所有service
		describeNamespaceServices(namespace)
	} else if len(args) == 2 {
		// 获取namespace和serviceName
		namespace, name, err := parseNameAndNamespace(args[1])

		if err != nil {
			fmt.Println("describeServiceHandler: parseNameAndNamespace error:", err)
			return
		}

		if namespace == "" || name == "" {
			fmt.Println("describeServiceHandler: namespace or name is empty")
			fmt.Println("Use like: kubectl describe service [serviceNamespace]/[serviceName]")
			return
		}

		// 获取特定的service
		describeSpecificService(namespace, name)
	}
}

func describeNamespaceServices(namespace string) {
	url := stringutil.Replace(config.ServiceURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = config.API_Server_URL_Prefix + url

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

	// 格式化输出JSON到控制台
	indentedJSON, err := json.MarshalIndent(services, "", "  ")

	if err != nil {
		fmt.Println("JSON marshaling failed:", err)
		return
	}

	fmt.Println(string(indentedJSON))
}

func describeSpecificService(namespace, name string) {
	url := stringutil.Replace(config.ServiceSpecURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, name)
	url = config.API_Server_URL_Prefix + url

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

	// 格式化输出JSON到控制台
	indentedJSON, err := json.MarshalIndent(service, "", "  ")
	if err != nil {
		fmt.Println("JSON marshaling failed:", err)
		return
	}

	fmt.Println(string(indentedJSON))
}

// ==============================================
//
// describe job handler
//
// ==============================================
func describeJobHandler(cmd *cobra.Command, args []string) {
	if len(args) == 1 {
		// 尝试获取用户是否制定了namespace
		namespace, _ := cmd.Flags().GetString("namespace")

		if namespace == "" {
			namespace = config.DefaultNamespace
		}

		// 获取default namespace下的所有job
		describeNamespaceJobs(namespace)
	} else if len(args) == 2 {
		// 获取namespace和jobName
		namespace, name, err := parseNameAndNamespace(args[1])

		if err != nil {
			fmt.Println("describeJobHandler: parseNameAndNamespace error:", err)
			return
		}

		if namespace == "" || name == "" {
			fmt.Println("describeJobHandler: namespace or name is empty")
			fmt.Println("Use like: kubectl describe job [jobNamespace]/[jobName]")
			return
		}

		// 获取特定的job
		describeSpecificJob(namespace, name)
	}
}

func describeNamespaceJobs(namespace string) {
	url := stringutil.Replace(config.JobsURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = config.API_Server_URL_Prefix + url

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

	// 格式化输出JSON到控制台
	indentedJSON, err := json.MarshalIndent(jobs, "", "  ")

	if err != nil {
		fmt.Println("JSON marshaling failed:", err)
		return
	}

	fmt.Println(string(indentedJSON))
}

func describeSpecificJob(namespace, name string) {
	url := stringutil.Replace(config.JobSpecURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, name)
	url = config.API_Server_URL_Prefix + url

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

	// 格式化输出JSON到控制台
	indentedJSON, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		fmt.Println("JSON marshaling failed:", err)
		return
	}

	fmt.Println(string(indentedJSON))
}

// ==============================================
//
// describe replicaset handler
//
// ==============================================
func describeReplicasetHandler(cmd *cobra.Command, args []string) {
	if len(args) == 1 {
		// 尝试获取用户是否制定了namespace
		namespace, _ := cmd.Flags().GetString("namespace")

		if namespace == "" {
			namespace = config.DefaultNamespace
		}

		// 获取default namespace下的所有replicaset
		describeNamespaceReplicasets(namespace)
	} else if len(args) == 2 {
		// 获取namespace和replicasetName
		namespace, name, err := parseNameAndNamespace(args[1])

		if err != nil {
			fmt.Println("describeReplicasetHandler: parseNameAndNamespace error:", err)
			return
		}

		if namespace == "" || name == "" {
			fmt.Println("describeReplicasetHandler: namespace or name is empty")
			fmt.Println("Use like: kubectl describe replicaset [replicasetNamespace]/[replicasetName]")
			return
		}

		// 获取特定的replicaset
		describeSpecificReplicaset(namespace, name)
	} else {
		fmt.Println("describeReplicasetHandler: args error")

	}
}

func describeNamespaceReplicasets(namespace string) {
	url := stringutil.Replace(config.ReplicaSetsURL, config.URL_PARAM_NAMESPACE_PART, namespace)

	url = config.API_Server_URL_Prefix + url

	replicasets := []apiObject.ReplicaSetStore{}

	code, err := netrequest.GetRequestByTarget(url, &replicasets, "data")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if code != http.StatusOK {
		fmt.Println("getNamespaceReplicasets: code:", code)
		return
	}

}

func describeSpecificReplicaset(namespace, name string) {
	url := stringutil.Replace(config.ReplicaSetSpecURL, config.URL_PARAM_NAMESPACE_PART, namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, name)
	url = config.API_Server_URL_Prefix + url

	replicaset := &apiObject.ReplicaSetStore{}

	code, err := netrequest.GetRequestByTarget(url, replicaset, "data")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if code != http.StatusOK {
		fmt.Println("getSpecificReplicaset: code:", code)
		return
	}

}
