package apiObject

import (
	"strings"
)

type GpuJobSpec struct {
	NumProcess      int      `yaml:"numProcess"`
	NumTasksPerNode int      `yaml:"numTasksPerNode"`
	CpusPerTask     int      `yaml:"cpusPerTask"`
	NumGpus         int      `yaml:"numGpus"`
	CompileScripts  []string `yaml:"compileScripts"`
	RunScripts      []string `yaml:"runScripts"`
	Volume          string   `yaml:"volume"`
	WorkDir         string   `yaml:"workDir"`
	OutputFile      string   `yaml:"outputFile"`
	ErrorFile       string   `yaml:"errorFile"`
	Username        string   `yaml:"username"`
	Password        string   `yaml:"password"`
}

type GpuJob struct {
	Base `yaml:",inline"`
	Spec GpuJobSpec `yaml:"spec"`
}

func (gpu *GpuJob) Namespace() string {
	return gpu.Metadata.Namespace
}

func (gpu *GpuJob) Name() string {
	return gpu.Metadata.Name
}

func (gpu *GpuJob) UID() types.UID {
	return gpu.Metadata.UID
}

func (gpu *GpuJob) Volume() string {
	return gpu.Spec.Volume
}

func (gpu *GpuJob) OutputFile() string {
	return gpu.Spec.OutputFile
}

func (gpu *GpuJob) ErrorFile() string {
	return gpu.Spec.ErrorFile
}

func (gpu *GpuJob) NumProcess() int {
	return gpu.Spec.NumProcess
}

func (gpu *GpuJob) NumTasksPerNode() int {
	return gpu.Spec.NumTasksPerNode
}

func (gpu *GpuJob) CpusPerTask() int {
	return gpu.Spec.CpusPerTask
}

func (gpu *GpuJob) NumGpus() int {
	return gpu.Spec.NumGpus
}

func (gpu *GpuJob) RunScripts() string {
	return strings.Join(gpu.Spec.RunScripts, ";")
}

func (gpu *GpuJob) CompileScripts() string {
	return strings.Join(gpu.Spec.CompileScripts, ";")
}

func (gpu *GpuJob) Username() string {
	return gpu.Spec.Username
}

func (gpu *GpuJob) Password() string {
	return gpu.Spec.Password
}

func (gpu *GpuJob) WorkDir() string {
	return gpu.Spec.WorkDir
}
