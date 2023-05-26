package apiObject

type Function struct {
	Basic `yaml:",inline" json:",inline"`
	Spec  FunctionSpec `yaml:"spec" json:"spec"`
}

// 用户需要上传zip文件，里面需要有相关的python文件
// python文件的入口文件应该是: func.py 函数名字是main
type FunctionSpec struct {
	UserUploadFile     []byte `yaml:"userUploadFile" json:"userUploadFile"`
	UserUploadFilePath string `yaml:"userUploadFilePath" json:"userUploadFilePath"`
}
