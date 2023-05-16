package sshclient

import "os"

var (
	SSHUserName = ""
	SSHPassword = ""
	SSHPort     = 22
)

const (
	SSH_Login_URL = "pilogin.hpc.sjtu.edu.cn"
	SSH_DATA_URL  = "data.hpc.sjtu.edu.cn"
)

func init() {
	// 从环境变量中读取用户名和密码
	SSHUserName = os.Getenv("GPU_SSH_USERNAME") // 从环境变量中读取用户名
	SSHPassword = os.Getenv("GPU_SSH_PASSWORD") // 从环境变量中读取密码

}
