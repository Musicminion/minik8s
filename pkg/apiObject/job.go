package apiObject

type Job struct {
	Basic `yaml:",inline"`
	Spec  JobSpec `yaml:"spec"`
}

type JobSpec struct {
	OutputFile string `yaml:"outputFile"`
	ErrorFile  string `yaml:"errorFile"`
}
