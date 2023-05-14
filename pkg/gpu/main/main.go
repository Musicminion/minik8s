package main

import (
	// "minik8s/pkg/gpu/gpu"
	"miniK8S/pkg/gpu/gp"
)

var (
	args = gpu.JobArgs{
		JobName:         "gpu-job",
		Output:          "output",
		Error:           "error",
		WorkDir:         "",
		NumProcess:      1,
		NumTasksPerNode: 1,
		CpusPerTask:     1,
		GpuResources:    "gpu:1",
		CompileScripts:  "",
		RunScripts:      "",
		Username:        "stu1638",
		Password:        "9Hi#GOjX",
	}
)

func main() {
	// 读取并设定参数
	// flag.StringVar(&args.JobName, "job-name", "gpu-job", "gpu job name")
	// flag.StringVar(&args.Output, "output", "output", "output filename")
	// flag.StringVar(&args.Error, "error", "error", "err filename")
	// flag.StringVar(&args.WorkDir, "workdir", "", "work directory")
	// flag.IntVar(&args.NumProcess, "process", 1, "number of processes(cpus)")
	// flag.IntVar(&args.NumTasksPerNode, "ntasks-per-node", 1, "number of tasks per node")
	// flag.IntVar(&args.CpusPerTask, "cpus-per-task", 1, "number of cpus per task")
	// flag.StringVar(&args.GpuResources, "gres", "gpu:1", "gpu resources")
	// flag.StringVar(&args.CompileScripts, "compile", "", "compile scripts")
	// flag.StringVar(&args.RunScripts, "run", "", "run scripts")
	// flag.StringVar(&args.Username, "username", "", "username")
	// flag.StringVar(&args.Password, "password", "", "password")
	// flag.Parse()
	server := gpu.NewServer(args, gpu.DefaultJobURL)
	server.Run()
}