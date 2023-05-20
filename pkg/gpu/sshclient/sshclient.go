package sshclient

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/melbahja/goph"
)

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

	// 上传文件夹
	UploadDir(localPath string, remotePath string) error
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
	fmt.Println("remotePath:", remotePath)
	fmt.Println("localPath:", localPath)
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

func (sc *sshClient) UploadDir(localPath string, remotePath string) error {
	err := filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 检查是否为文件夹
		if info.IsDir() {
			return nil
		}

		// 获取相对路径
		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			return err
		}

		// 构建远程文件路径
		remotePath := filepath.Join(remotePath, relPath)

		// 使用 SSH 客户端上传文件
		err = sc.Client.Upload(path, remotePath)

		if err != nil {
			fmt.Println("上传文件失败:", err)
		} else {
			fmt.Println("文件上传成功:", remotePath)
		}

		return nil
	})

	return err
}
