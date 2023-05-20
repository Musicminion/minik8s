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

func (js *JobStore) GetJobName() string {
	return js.Metadata.Name
}

func (js *JobStore) GetJobNamespace() string {
	return js.Metadata.Namespace
}

func (js *JobStore) GetJobUUID() string {
	return js.Metadata.UUID
}

func (j *Job) GetJobName() string {
	return j.Basic.Metadata.Name
}

func (j *Job) GetJobNamespace() string {
	return j.Basic.Metadata.Namespace
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

func (js *JobStore) ToJob() *Job {
	return &Job{
		Basic: js.Basic,
		Spec:  js.Spec,
	}
}

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
