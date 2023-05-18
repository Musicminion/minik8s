package jobserver

var (
	AcceptFileSuffix = []string{".cu"}
)

type JobServerConfig struct {
	Username    string   `yaml:"username" json:"username"`
	Password    string   `yaml:"password" json:"password"`
	WorkDir     string   `yaml:"workDir" json:"workDir"`
	CompileCmds []string `yaml:"compileCmds" json:"compileCmds"`
}

func NewJobServerConfig() *JobServerConfig {
	return &JobServerConfig{
		Username: "",
		Password: "",
	}
}
