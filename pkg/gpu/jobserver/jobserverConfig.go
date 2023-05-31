package jobserver

import (
	"errors"
	"time"
)

var (
	AcceptFileSuffix = []string{".cu"}
)

// 脚本相关的文件里面的填充信息
const (
	SBATCH_HEADER     = `#!/bin/bash`
	SBATCH_JOB_Name   = `#SBATCH --job-name=%s`
	SBATCH_OUTPUT     = `#SBATCH --output=%s`
	SBATCH_ERROR      = `#SBATCH --error=%s`
	SBATCH_PARTITION  = `#SBATCH --partition=%s`
	SBATCH_GPUS       = `#SBATCH --gres=gpu:%d`      // GPU数量
	SBATCH_TOTAL_CPUS = `#SBATCH --ntasks=%d`        // 总共的CPU数量
	SBATCH_NODE_CPUS  = `#SBATCH --cpus-per-task=%d` // 每个节点的CPU数量
	SBATCH_NODE_NUMS  = `#SBATCH -N %d`              // 节点数量

	SBATCH_NEXT_LINE = "\n"

	// SBATCH_TIME      = `#SBATCH --time=%s`
	// SBATCH_MEM        = `#SBATCH --mem=%d`
	// SBATCH_MAIL       = `#SBATCH --mail-type=%s`
	// SBATCH_MAIL_USER  = `#SBATCH --mail-user=%s`
	// SBATCH_NODES      = `#SBATCH --nodes=%d`         // 节点数量

	// .slurm
	SBATCH_SUFFIX = ".slurm"

	// 提交文件的命令
	SBATCH_SUBMIT                  = "sbatch -D %s %s"
	SBATCH_ACCMPLISH               = "sacct -j %s"
	SBATCH_ACCMPLISH_FliterHead    = "tail -n +3"
	SBATCH_ACCMPLISH_FliterContent = "awk '{print $1, $2, $3, $4, $5, $6, $7}'"
)

var (
	UploadFileDelay = 10 * time.Second
	ExecutorJob_Delay  = 0 * time.Second
	ExecutorJob_Period = []time.Duration{10 * time.Second}
	ExecutorJob_IfLoop = true
)

type JobServerConfig struct {
	Username      string   `yaml:"username" json:"username"`           // 用户名
	Password      string   `yaml:"password" json:"password"`           // 密码
	WorkDir       string   `yaml:"workDir" json:"workDir"`             // 工作目录(用的是绝对路径)
	RemoteWorkDir string   `yaml:"remoteWorkDir" json:"remoteWorkDir"` // 远程工作目录(用的是相对路径)
	CompileCmds   []string `yaml:"compileCmds" json:"compileCmds"`     // 编译命令
	RunCmds       []string `yaml:"runCmds" json:"runCmds"`             // 运行命令
	JobName       string   `yaml:"jobName" json:"jobName"`             // 任务的名字
	JobNamespace  string   `yaml:"jobNamespace" json:"jobNamespace"`   // 任务的命名空间
	JobUUID       string   `yaml:"jobUUID" json:"jobUUID"`             // 任务的UUID
	OutputFile    string   `yaml:"outputFile" json:"outputFile"`       // 输出文件
	ErrorFile     string   `yaml:"errorFile" json:"errorFile"`         // 错误文件
	Partition     string   `yaml:"partition" json:"partition"`         // 分区
	GPUNums       int      `yaml:"gpuNums" json:"gpuNums"`             // GPU数量
	CPUNums       int      `yaml:"cpuNums" json:"cpuNums"`             // 总共的CPU数量
	NodeCPUNums   int      `yaml:"nodeCPUNums" json:"nodeCPUNums"`     // 每个节点的CPU数量

	// 和API Server通讯的地址
	APIServerAddr string `yaml:"apiServerAddr" json:"apiServerAddr"`
	// 以下参数已废弃，待删除
	// // 因为CompileCmds 和 RunCmds 里面的命令都是从命令行解析的，解析的格式是按照;分割的
	// CmdArgCompileCmds string `yaml:"cmdArgCompileCmds" json:"cmdArgCompileCmds"`
	// CmdArgRunCmds     string `yaml:"cmdArgRunCmds" json:"cmdArgRunCmds"`
}

func NewJobServerConfig() *JobServerConfig {
	return &JobServerConfig{
		Username: "",
		Password: "",
	}
}

// [stu1638@pilogin5 ~]$ sacct
// JobID           JobName  Partition    Account  AllocCPUS      State ExitCode
// ------------ ---------- ---------- ---------- ---------- ---------- --------
type JobCompleteStatus struct {
	JobID     string `json:"jobID" yaml:"jobID"`
	JobName   string `json:"jobName" yaml:"jobName"`
	Partition string `json:"partition" yaml:"partition"`
	Account   string `json:"account" yaml:"account"`
	AllocCPUS string `json:"allocCPUS" yaml:"allocCPUS"`
	State     string `json:"state" yaml:"state"`
	ExitCode  string `json:"exitCode" yaml:"exitCode"`
}

func NewJobCompleteStatus(data []string) (*JobCompleteStatus, error) {
	if len(data) != 7 {
		return nil, errors.New("CheckStatus failed, for result format error")
	}

	// 解析JobID
	return &JobCompleteStatus{
		JobID:     data[0],
		JobName:   data[1],
		Partition: data[2],
		Account:   data[3],
		AllocCPUS: data[4],
		State:     data[5],
		ExitCode:  data[6],
	}, nil
}
