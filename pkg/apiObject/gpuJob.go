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
	Basic `yaml:",inline"`
	Spec  GpuJobSpec `yaml:"spec"`
}

func (gpu *GpuJob) GetNamespace() string {
	return gpu.Metadata.Namespace
}

func (gpu *GpuJob) GetName() string {
	return gpu.Metadata.Name
}

func (gpu *GpuJob) GetGetUUID() string {
	return gpu.Metadata.UUID
}

func (gpu *GpuJob) GetVolume() string {
	return gpu.Spec.Volume
}

func (gpu *GpuJob) GetOutputFile() string {
	return gpu.Spec.OutputFile
}

func (gpu *GpuJob) GetErrorFile() string {
	return gpu.Spec.ErrorFile
}

func (gpu *GpuJob) GetNumProcess() int {
	return gpu.Spec.NumProcess
}

func (gpu *GpuJob) GetNumTasksPerNode() int {
	return gpu.Spec.NumTasksPerNode
}

func (gpu *GpuJob) GetCpusPerTask() int {
	return gpu.Spec.CpusPerTask
}

func (gpu *GpuJob) GetNumGpus() int {
	return gpu.Spec.NumGpus
}

func (gpu *GpuJob) GetRunScripts() string {
	return strings.Join(gpu.Spec.RunScripts, ";")
}

func (gpu *GpuJob) GetCompileScripts() string {
	return strings.Join(gpu.Spec.CompileScripts, ";")
}

func (gpu *GpuJob) GetUsername() string {
	return gpu.Spec.Username
}

func (gpu *GpuJob) GetPassword() string {
	return gpu.Spec.Password
}

func (gpu *GpuJob) GetWorkDir() string {
	return gpu.Spec.WorkDir
}
