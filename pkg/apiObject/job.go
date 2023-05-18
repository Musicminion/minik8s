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
	NTasks          int      `yaml:"nTasks" json:"nTasks"`                   // 进程总数
	NTasksPerNode   int      `yaml:"nTasksPerNode" json:"nTasksPerNode"`     // 每个节点请求的任务数
	SubmitDirectory string   `yaml:"submitDirectory" json:"submitDirectory"` // 提交目录
	SubmitHost      string   `yaml:"submitHost" json:"submitHost"`           // 提交主机
	CompileCommands []string `yaml:"compileCommands" json:"compileCommands"` // 编译命令
	RunCommands     []string `yaml:"runCommands" json:"runCommands"`         // 运行命令
	OutputFile      string   `yaml:"outputFile" json:"outputFile"`           // 输出文件
	ErrorFile       string   `yaml:"errorFile" json:"errorFile"`             // 错误文件
	JobUsername     string   `yaml:"username" json:"username"`               // 用户名
	JobPassword     string   `yaml:"password" json:"password"`               // 密码
}

type JobStatus struct {
	State      string    `yaml:"state" json:"state"`
	UpdateTime time.Time `yaml:"updateTime" json:"updateTime"`
}

type JobStore struct {
	Basic `yaml:",inline" json:",inline"`
	Spec  JobSpec `yaml:"spec" json:"spec"`
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
