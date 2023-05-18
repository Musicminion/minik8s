package jobserver

import (
	"io/fs"
	"miniK8s/pkg/gpu/sshclient"
	"path/filepath"
)

type JobServer struct {
	conf      *JobServerConfig
	sshClient sshclient.SSHClient
}

func NewJobServer(conf *JobServerConfig) (*JobServer, error) {
	client, err := sshclient.NewSSHClient(conf.Username, conf.Password)
	if err != nil {
		return nil, err
	}

	return &JobServer{
		conf:      conf,
		sshClient: client,
	}, nil
}

func (js *JobServer) FindJobFiles() []string {
	filesPath := make([]string, 0)

	checkFun := func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			fileName := d.Name()

			// 遍历AcceptFileSuffix
			for _, suffix := range AcceptFileSuffix {
				if filepath.Ext(fileName) == suffix {
					filesPath = append(filesPath, fileName)
					break
				}
			}

		}
		return nil
	}

	_ = filepath.WalkDir("", checkFun)

	return filesPath
}

func (js *JobServer) UpdateJobFiles(filePaths []string) {
	for _, filePath := range filePaths {
		js.sshClient.UploadFile(filePath, js.conf.WorkDir)
	}
}

func (js *JobServer) CompileJobFiles() {
	js.sshClient.RunCmds(js.conf.CompileCmds)
}

// 运行Job开始之前的准备工作
func (js *JobServer) SetupJob() error {
	// 清空工作目录
	js.sshClient.RemoveDirectory(js.conf.WorkDir)
	js.sshClient.MakeDirectory(js.conf.WorkDir)

	// 获取所有需要上传的文件
	uploadFiles := js.FindJobFiles()
	// 上传文件
	js.UpdateJobFiles(uploadFiles)

	// 编译文件
	js.CompileJobFiles()

	return nil
}

func (js *JobServer) Run() {

}
