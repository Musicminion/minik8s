package jobserver

import (
	"errors"
	"fmt"
	"io/fs"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/gpu/sshclient"
	"miniK8s/util/executor"
	"path/filepath"
	"strings"
	"time"
)

type JobServer struct {
	conf      *JobServerConfig
	sshClient sshclient.SSHClient
}

func NewJobServer(conf *JobServerConfig) (*JobServer, error) {
	client, err := sshclient.NewSSHClient(conf.Username, conf.Password)
	if err != nil {
		return nil, err
	}

	return &JobServer{
		conf:      conf,
		sshClient: client,
	}, nil
}

func (js *JobServer) FindJobFiles() []string {
	filesPath := make([]string, 0)

	checkFun := func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			fileName := d.Name()

			// 遍历AcceptFileSuffix
			for _, suffix := range AcceptFileSuffix {
				if filepath.Ext(fileName) == suffix {
					filesPath = append(filesPath, fileName)
					break
				}
			}

		}
		return nil
	}

	_ = filepath.WalkDir("", checkFun)

	return filesPath
}

func (js *JobServer) UpdateJobFiles(filePaths []string) {
	for _, filePath := range filePaths {
		js.sshClient.UploadFile(filePath, js.conf.WorkDir)
	}
}

func (js *JobServer) CompileJobFiles() {
	js.sshClient.RunCmds(js.conf.CompileCmds)
}

func (js *JobServer) CompactRunCmd() string {
	result := ""
	for _, cmd := range js.conf.RunCmds {
		result += cmd + "\n"
	}
	return result
}

func (js *JobServer) CompactJobFiles() string {
	var fileContent string
	fileContent += SBATCH_HEADER + SBATCH_NEXT_LINE
	fileContent += fmt.Sprintf(SBATCH_JOB_Name, js.conf.JobName) + SBATCH_NEXT_LINE
	fileContent += fmt.Sprintf(SBATCH_OUTPUT, js.conf.OutputFile) + SBATCH_NEXT_LINE
	fileContent += fmt.Sprintf(SBATCH_ERROR, js.conf.ErrorFile) + SBATCH_NEXT_LINE
	fileContent += fmt.Sprintf(SBATCH_PARTITION, js.conf.Partition) + SBATCH_NEXT_LINE
	fileContent += fmt.Sprintf(SBATCH_GPUS, js.conf.GPUNums) + SBATCH_NEXT_LINE
	fileContent += fmt.Sprintf(SBATCH_TOTAL_CPUS, js.conf.CPUNums) + SBATCH_NEXT_LINE
	fileContent += fmt.Sprintf(SBATCH_NODE_CPUS, js.conf.NodeCPUNums) + SBATCH_NEXT_LINE
	fileContent += SBATCH_NEXT_LINE
	fileContent += js.CompactRunCmd() + SBATCH_NEXT_LINE
	return fileContent
}

func (js *JobServer) CompactJobPath() string {
	path := js.conf.WorkDir + js.conf.JobName + SBATCH_SUFFIX
	return path
}

// 运行Job开始之前的准备工作
func (js *JobServer) SetupJob() error {
	// 清空工作目录
	_, err := js.sshClient.RemoveDirectory(js.conf.WorkDir)
	if err != nil {
		fmt.Println("RemoveDirectory failed, for err" + err.Error())
		return err
	}
	_, err = js.sshClient.MakeDirectory(js.conf.WorkDir)

	if err != nil {
		fmt.Println("MakeDirectory failed, for err" + err.Error())
		return err
	}

	// 获取所有需要上传的文件
	uploadFiles := js.FindJobFiles()
	// 上传文件
	js.UpdateJobFiles(uploadFiles)

	// 编译文件
	js.CompileJobFiles()

	// 获取Job文件内容
	jobFileContent := js.CompactJobFiles()

	// 上传Job文件
	_, err = js.sshClient.WriteFile(jobFileContent, js.CompactJobPath())

	if err != nil {
		fmt.Println("WriteFile failed, for err" + err.Error())
		return err
	}

	return nil
}

// 返回任务的ID
func (js *JobServer) SubmitJob() (string, error) {
	// 提交Job的命令
	command := fmt.Sprintf(SBATCH_SUBMIT, js.CompactJobPath())
	result, err := js.sshClient.RunCmd(command)

	if err != nil {
		fmt.Println("RunCmd failed, for err" + err.Error())
		return "", err
	}

	// 尝试解析JobID
	var jobID string
	n, err := fmt.Sscanf(result, "Submitted batch job %s", &jobID)

	if n < 1 || n > 1 || err != nil {
		fmt.Println("Sscanf failed, for err" + err.Error())
		return "", err
	}

	return jobID, nil
}

// 检查任务状态
func (js *JobServer) CheckStatus(jobID string) (*JobCompleteStatus, error) {
	command := fmt.Sprintf(SBATCH_ACCMPLISH, jobID) + "|" + SBATCH_ACCMPLISH_FliterHead + "|" + SBATCH_ACCMPLISH_FliterContent

	result, err := js.sshClient.RunCmd(command)
	if err != nil {
		fmt.Println("RunCmd failed, for err" + err.Error())
		return nil, err
	}

	resultLines := strings.Split(result, "\n")
	if len(resultLines) > 0 {
		result = resultLines[0]

		// 解析结果
		data := strings.Split(result, " ")
		if len(data) != 7 {
			return nil, errors.New("CheckStatus failed, for result format error")
		}

		return NewJobCompleteStatus(data)
	}

	// 没有结果
	return nil, nil
}

// 更新任务的状态
// 返回1表示任务完成，返回0表示任务未完成，返回-1表示任务出错
func (js *JobServer) CheckAndUpdateJobStatus(jobID string) int {
	// 获取任务状态
	status, err := js.CheckStatus(jobID)
	if err != nil {
		fmt.Println("CheckStatus failed, for err" + err.Error())
		return -1
	}

	// 更新任务状态
	if status != nil {
		jobStatus := &apiObject.JobStatus{
			JobID:      jobID,
			Partition:  status.Partition,
			Account:    status.Account,
			AllocCPUS:  status.AllocCPUS,
			State:      status.State,
			ExitCode:   status.ExitCode,
			UpdateTime: time.Now(),
		}

		// 更新任务状态 TODO 写入到apiServer【TODO】
		fmt.Println(jobStatus)
		return 1
	}
	return 0
}

// 自旋等待
func (js *JobServer) Spin() {
	// 进入睡眠死循环
	<-make(chan struct{})
}

func (js *JobServer) Run() {
	err := js.SetupJob()
	if err != nil {
		fmt.Println("SetupJob failed, for err" + err.Error())
		return
	}

	// 启动Job
	jobID, err := js.SubmitJob()
	if err != nil {
		fmt.Println("SubmitJob failed, for err" + err.Error())
		return
	}
	fmt.Println("JobID: " + jobID)

	periodCheck := func() bool {
		switch js.CheckAndUpdateJobStatus(jobID) {
		case 0:
			return false // 任务未完成，继续循环
		case 1:
			return true // 任务完成，结束循环
		case -1:
			return true // 任务出错，结束循环
		}
		return false
	}

	// 启动定时器
	executor.ConditionPeriod(ExecutorJob_Delay, ExecutorJob_Period, periodCheck, ExecutorJob_IfLoop)
	// 完成任务，等待
	// 自旋等待
	js.Spin()
}
