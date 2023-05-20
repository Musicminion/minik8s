package main

import (
	"errors"
	"flag"
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/pkg/gpu/jobserver"
	netrequest "miniK8s/util/netRequest"
	"miniK8s/util/stringutil"
	"miniK8s/util/zip"
	"net/http"
	"os"
	"strings"
)

type mainArg struct {
	JobName       string // 任务的名字
	JobNamespace  string // 任务的命名空间
	APIServerAddr string // 与API Server通讯的地址 默认值为 http://127.0.0.1:8090
}

var (
	args mainArg
)

const (
	Default_API_Server_Addr = "http://127.0.0.1:8090"
	JobDirectory            = "/job/"
	JobResultDirectory      = JobDirectory + "jobresult"
)

// 通过API Server获取任务的配置信息
// 通过API Server获取任务的配置信息
func getJobFromAPIServer() (*apiObject.JobStore, error) {
	jobURL := stringutil.Replace(config.JobSpecURL, config.URL_PARAM_NAMESPACE_PART, args.JobNamespace)
	jobURL = stringutil.Replace(jobURL, config.URL_PARAM_NAME_PART, args.JobName)
	jobURL = args.APIServerAddr + jobURL

	// 通过http请求获取任务的配置信息
	job := &apiObject.JobStore{}
	code, err := netrequest.GetRequestByTarget(jobURL, job, "data")

	if err != nil {
		return nil, err
	}

	if code != http.StatusOK {
		return nil, errors.New("get job info from api server failed")
	}

	return job, nil
}

// 下载1个任务的文件，存贮在本地，并且解压
func getJobFileFromAPIServer(job *apiObject.JobStore, conf *jobserver.JobServerConfig) error {
	// 1. 获取文件的URL
	fileURL := stringutil.Replace(config.JobFileSpecURL, config.URL_PARAM_NAMESPACE_PART, args.JobNamespace)
	fileURL = stringutil.Replace(fileURL, config.URL_PARAM_NAME_PART, args.JobName)
	fileURL = args.APIServerAddr + fileURL

	// 2. 下载文件
	jobFile := &apiObject.JobFile{}

	fmt.Println("fileURL:", fileURL)
	code, err := netrequest.GetRequestByTarget(fileURL, jobFile, "data")

	if err != nil {
		return err
	}

	if code != http.StatusOK {
		return errors.New("get job file from api server failed")
	}

	if len(jobFile.UserUploadFile) == 0 {
		return errors.New("job file is empty")
	}

	jobZipfileName := "job-" + jobFile.GetJobFileUUID()
	jobZipfileNameFull := jobZipfileName + ".zip"

	// 清空JobDirectory
	err = os.RemoveAll(JobDirectory)

	if err != nil {
		return err
	}

	// 确保目录存在
	err = os.MkdirAll(JobDirectory, os.ModePerm)

	if err != nil {
		return err
	}

	// 3. 转化文件
	err = zip.ConvertBytesToZip(jobFile.UserUploadFile, JobDirectory+jobZipfileNameFull)

	if err != nil {
		return err
	}

	// 4. 解压文件
	err = zip.DecompressZip(JobDirectory+jobZipfileNameFull, conf.WorkDir)

	if err != nil {
		return err
	}

	// 解析job里面的submit的文件夹名
	submitDir := job.Spec.SubmitDirectory

	// 从后往前找到第一个/的位置
	index := strings.LastIndex(submitDir, "/")
	if index == -1 {
		return errors.New("submit directory is invalid")
	}
	// 截取submit的文件夹名
	submitDir = submitDir[index+1:]

	if len(submitDir) == 0 {
		return errors.New("submit directory is invalid")
	}
	// 修改文件夹的名字
	err = os.Rename(conf.WorkDir+submitDir, conf.WorkDir+jobZipfileName)

	if err != nil {
		return err
	}

	conf.WorkDir = conf.WorkDir + jobZipfileName
	conf.RemoteWorkDir = conf.RemoteWorkDir + jobZipfileName
	return nil
}

// 准备任务的配置信息
func prepareJobConfig() (*jobserver.JobServerConfig, error) {
	conf := jobserver.NewJobServerConfig()
	jobInfo, err := getJobFromAPIServer()

	if err != nil {
		return nil, err
	}

	if jobInfo == nil {
		return nil, fmt.Errorf("JobName %s is not exist", args.JobName)
	}

	// =================== 根据jobInfo的信息去准备任务的配置信息 ===================
	// 1. 任务的名字
	conf.JobName = jobInfo.Metadata.Name
	// 2. 任务的命名空间
	conf.JobNamespace = jobInfo.Metadata.Namespace
	// 3. CPU的数量
	conf.CPUNums = jobInfo.Spec.NTasks
	// 4. GPU的数量
	conf.GPUNums = jobInfo.Spec.GPUNums
	// 5. 任务的output文件
	conf.OutputFile = jobInfo.Spec.OutputFile
	// 6. 任务的error文件
	conf.ErrorFile = jobInfo.Spec.ErrorFile
	// 7. 任务的工作目录
	conf.WorkDir = JobDirectory
	// 8. 任务的编译命令
	conf.CompileCmds = jobInfo.Spec.CompileCommands
	// 9. 任务的运行命令
	conf.RunCmds = jobInfo.Spec.RunCommands
	// 10. 登录的用户名
	conf.Username = jobInfo.Spec.JobUsername
	// 11. 登录的密码
	conf.Password = jobInfo.Spec.JobPassword
	// 12. 任务的Partitions
	conf.Partition = jobInfo.Spec.JobPartition
	// 13. 任务每个节点的CPU数量
	conf.NodeCPUNums = jobInfo.Spec.NTasksPerNode
	// 14. API Server的地址
	conf.APIServerAddr = args.APIServerAddr
	// 15. 任务的ID
	conf.JobUUID = jobInfo.Metadata.UUID
	// 16. 远程的工作目录
	conf.RemoteWorkDir = "job/"

	// ==========================================================================

	// 这一步会修改conf.WorkDir的值
	err = getJobFileFromAPIServer(jobInfo, conf)

	if err != nil {
		return nil, err
	}

	return conf, nil
}

// 主函数体的接受的参数尽可能的简介，然后他会主动的去请求api server
// 根据请求的结果去决定任务的执行
// 程序的使用方法： -jobName YourJobName -jobNamespace YourJobNamespace -apiServerAddr YourAPIServerAddr
// YourAPIServerAddr: http://192.168.126.130:8090
func main() {
	// 第一个参数 指针、第二个参数的名字 第三个参数默认值 第四个参数的描述帮助信息
	flag.StringVar(&args.JobName, "jobName", "", "-jobName YourJobName")
	flag.StringVar(&args.JobNamespace, "jobNamespace", "", "-jobNamespace YourJobNamespace")
	flag.StringVar(&args.APIServerAddr, "apiServerAddr", Default_API_Server_Addr, "-apiServerAddr YourAPIServerAddr")
	flag.Parse()

	if args.JobName == "" || args.JobNamespace == "" {
		fmt.Println("JobName or JobNamespace is empty, usage is: -jobName YourJobName -jobNamespace YourJobNamespace")
		return
	}

	fmt.Printf("JobName is %s, JobNamespace is %s \n", args.JobName, args.JobNamespace)

	// 和APi Server通讯，准备任务的文件
	conf, err := prepareJobConfig()

	if err != nil {
		fmt.Printf("prepareJobConfig error: %s \n", err)
		return
	}

	if conf == nil {
		fmt.Println("prepareJobConfig conf is nil")
		return
	}

	// 通过jobserver去执行任务
	jobserver, err := jobserver.NewJobServer(conf)

	if err != nil {
		fmt.Printf("NewJobServer error: %s", err)
		return
	}

	jobserver.Run()

	// TODO
	// jobserver, err := jobserver.NewJobServer(conf)
	// if err != nil {
	// 	panic(err)
	// }
	// if jobserver == nil {
	// 	panic("jobserver is nil")
	// }
	// jobserver.NewJobServer(jobserver.NewJobServerConfig())
	// time.Sleep(1000 * time.Second)

}
