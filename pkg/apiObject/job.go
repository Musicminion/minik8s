package apiObject

import "time"

type Job struct {
	Basic `yaml:",inline"`
	Spec  JobSpec `yaml:"spec"`
}

// 详细解释参考：https://docs.hpc.sjtu.edu.cn/job/slurm.html
// https://docs.hpc.sjtu.edu.cn/job/slurm.html
type JobSpec struct {
	JobPartition    string   `yaml:"partition" json:"partition"`             // 分区
	NTasks          int      `yaml:"nTasks" json:"nTasks"`                   // 进程总数，也是CPU核心数
	NTasksPerNode   int      `yaml:"nTasksPerNode" json:"nTasksPerNode"`     // 每个节点请求的任务数
	SubmitDirectory string   `yaml:"submitDirectory" json:"submitDirectory"` // 提交目录
	CompileCommands []string `yaml:"compileCommands" json:"compileCommands"` // 编译命令
	RunCommands     []string `yaml:"runCommands" json:"runCommands"`         // 运行命令
	OutputFile      string   `yaml:"outputFile" json:"outputFile"`           // 输出文件
	ErrorFile       string   `yaml:"errorFile" json:"errorFile"`             // 错误文件
	JobUsername     string   `yaml:"username" json:"username"`               // 用户名
	JobPassword     string   `yaml:"password" json:"password"`               // 密码
	GPUNums         int      `yaml:"gpuNums" json:"gpuNums"`                 // GPU数目

	// SubmitHost      string   `yaml:"submitHost" json:"submitHost"`           // 提交主机
}

// 任务状态，参考：https://docs.hpc.sjtu.edu.cn/job/slurm.html
type JobStatus struct {
	JobID      string    `yaml:"jobID" json:"jobID"`           // 任务ID，这个是slurm返回的ID
	Partition  string    `yaml:"partition" json:"partition"`   // 分区
	Account    string    `yaml:"account" json:"account"`       // 账户
	AllocCPUS  string    `yaml:"allocCPUS" json:"allocCPUS"`   // 分配CPU数
	State      string    `yaml:"state" json:"state"`           // 任务状态
	ExitCode   string    `yaml:"exitCode" json:"exitCode"`     // 退出码
	Output     []string  `yaml:"output" json:"output"`         // 输出的内容
	Error      []string  `yaml:"error" json:"error"`           // 错误的内容
	UpdateTime time.Time `yaml:"updateTime" json:"updateTime"` // 更新时间
}

type JobStore struct {
	Basic  `yaml:",inline" json:",inline"`
	Spec   JobSpec   `yaml:"spec" json:"spec"`
	Status JobStatus `yaml:"status" json:"status"`
}

func (j *Job) GetJobUUID() string {
	return j.Basic.Metadata.UUID
}

func (j *Job) ToJobStore() *JobStore {
	return &JobStore{
		Basic: j.Basic,
		Spec:  j.Spec,
	}
}

func (js *JobStore) GetJobName() string {
	return js.Metadata.Name
}

func (js *JobStore) GetJobNamespace() string {
	return js.Metadata.Namespace
}

func (js *JobStore) GetJobUUID() string {
	return js.Metadata.UUID
}

func (js *JobStore) ToJob() *Job {
	return &Job{
		Basic: js.Basic,
		Spec:  js.Spec,
	}
}

// 以下函数用来实现apiObject.Object接口
func (j *Job) GetObjectKind() string {
	return j.Kind
}

func (j *Job) GetObjectName() string {
	return j.Metadata.Name
}

func (j *Job) GetObjectNamespace() string {
	return j.Metadata.Namespace
}

///////////////////////// 以下是JobFile相关的内容 /////////////////////////

type JobFile struct {
	Basic          `yaml:",inline" json:",inline"`
	UserUploadFile []byte `yaml:"userUploadFile" json:"userUploadFile"` // 用户上传的文件 zip文件
	OutputFile     []byte `yaml:"outputFile" json:"outputFile"`         // 输出文件，执行的结果
	ErrorFile      []byte `yaml:"errorFile" json:"errorFile"`           // 错误文件，执行的结果
}

func (jf *JobFile) GetJobName() string {
	return jf.Metadata.Name
}

func (jf *JobFile) GetJobNamespace() string {
	return jf.Metadata.Namespace
}

func (jf *JobFile) GetJobFileUUID() string {
	return jf.Metadata.UUID
}

// 任务状态

const (
	// BF BOOT_FAIL       Job terminated due to launch failure, typically due to a hardware failure (e.g. unable to boot the node or block and the job can not be requeued).
	JobState_BOOT_FAIL = "BOOT_FAIL" // 任务启动失败
	//  CA CANCELLED       Job was explicitly cancelled by the user or system administrator. The job may or may not have been initiated.
	JobState_CANCELLED = "CANCELLED" // 任务被取消
	// CL COMPLETED       Job has terminated all processes on all nodes with an exit code of zero.
	JobState_COMPLETED = "COMPLETED" // 任务完成
	// DL  DEADLINE        Job terminated on deadline.
	JobState_DEADLINE = "DEADLINE" // 任务超时
	// F  FAILED          Job terminated with non-zero exit code or other failure condition.
	JobState_FAILED = "FAILED" // 任务失败
	// NF NODE_FAIL       Job terminated due to failure of one or more allocated nodes.
	JobState_NODE_FAIL = "NODE_FAIL" // 节点失败
	// OOM OUT_OF_MEMORY  Job experienced out of memory error.
	JobState_OUT_OF_MEMORY = "OUT_OF_MEMORY" // 内存不足
	// PD PENDING         Job is awaiting resource allocation.
	JobState_PENDING = "PENDING" // 任务等待资源分配
	// PR PREEMPTED       Job terminated due to preemption.
	JobState_PREEMPTED = "PREEMPTED" // 任务被抢占
	// R  RUNNING         Job currently has an allocation.
	JobState_RUNNING = "RUNNING" // 任务正在运行
	// S  SUSPENDED       Job has an allocation, but execution has been suspended.
	JobState_SUSPENDED = "SUSPENDED" // 任务被挂起
	// TO TIMEOUT         Job terminated upon reaching its time limit.
	JobState_TIMEOUT = "TIMEOUT" // 任务超时
	// CG COMPLETING      Job is in the process of completing. Some processes on some nodes may still be active.
	JobState_COMPLETING = "COMPLETING" // 任务正在完成
	// CD COMPLETED       Job has terminated all processes on all nodes.
	//  RV  REVOKED         Sibling was removed from cluster due to other cluster starting the job.
	JobState_REVOKED = "REVOKED" // 任务被撤销

)
