package sshclient

import "github.com/melbahja/goph"

type SSHClient interface {
	// 执行命令
	RunCmd(cmd string) (string, error)
	// 执行多条命令
	RunCmds(cmds []string) (string, error)

	// 切换目录
	ChangeDirectory(path string) (string, error)
	RemoveDirectory(path string) (string, error)
	MakeDirectory(path string) (string, error)

	// File相关的操作
	ReadFile(path string) (string, error)
	WriteFile(path string, content string) (string, error)
	AppendFile(path string, content string) (string, error)
	RemoveFile(path string) (string, error)

	// 上传文件
	UploadFile(localPath string, remotePath string) error
	DownloadFile(remotePath string, localPath string) error
}

type sshClient struct {
	Username string
	Password string
	Client   *goph.Client
}

func NewSSHClient(username, password string) (SSHClient, error) {
	// 创建一个新的SSH客户端
	cli, err := goph.NewUnknown(username, SSH_Login_URL, goph.Password(password))
	if err != nil {
		return nil, err
	}

	return &sshClient{
		Username: username,
		Password: password,
		Client:   cli,
	}, nil
}

func (sc *sshClient) ChangeDirectory(path string) (string, error) {
	out, err := sc.Client.Run(ChangeDirectoryCommand + path)
	return string(out), err
}

func (sc *sshClient) RemoveDirectory(path string) (string, error) {
	out, err := sc.Client.Run(RemoveDirectoryCommand + path)
	return string(out), err
}

func (sc *sshClient) MakeDirectory(path string) (string, error) {
	out, err := sc.Client.Run(MakeDirectoryCommand + path)
	return string(out), err
}

func (sc *sshClient) ReadFile(path string) (string, error) {
	out, err := sc.Client.Run(ReadFileCommand + path)
	return string(out), err
}

func (sc *sshClient) WriteFile(path string, content string) (string, error) {
	writeCmd := WriteFileCommand + "\"" + content + "\" > " + path
	out, err := sc.Client.Run(writeCmd)
	return string(out), err
}

func (sc *sshClient) AppendFile(path string, content string) (string, error) {
	appendCmd := AppendFileCommand + "\"" + content + "\" >> " + path
	out, err := sc.Client.Run(appendCmd)
	return string(out), err
}

func (sc *sshClient) RemoveFile(path string) (string, error) {
	out, err := sc.Client.Run(RemoveFileCommand + path)
	return string(out), err
}

func (sc *sshClient) RunCmd(cmd string) (string, error) {
	out, err := sc.Client.Run(cmd)
	return string(out), err
}

func (sc *sshClient) UploadFile(localPath string, remotePath string) error {
	// 上传文件
	err := sc.Client.Upload(localPath, remotePath)
	return err
}

func (sc *sshClient) DownloadFile(remotePath string, localPath string) error {
	// 下载文件
	err := sc.Client.Download(remotePath, localPath)
	return err
}

func (sc *sshClient) RunCmds(cmds []string) (string, error) {
	var out string
	var err error
	for _, cmd := range cmds {
		curOut, err := sc.Client.Run(cmd)
		if err != nil {
			return out, err
		}
		out += string(curOut)
	}
	return out, err
}
