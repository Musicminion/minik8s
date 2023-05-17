package sshclient

import (
	"os"
	"testing"
)

func TestEnv(t *testing.T) {
	t.Log("GPU_SSH_USERNAME: ", os.Getenv("GPU_SSH_USERNAME"))
	t.Log("GPU_SSH_PASSWORD: ", os.Getenv("GPU_SSH_PASSWORD"))
}

func TestNewSSHClient(t *testing.T) {
	// client, err := NewSSHClient(os.Getenv("GPU_SSH_USERNAME"), os.Getenv("GPU_SSH_PASSWORD"))

	// if err != nil {
	// 	t.Error(err)
	// }

	// if client == nil {
	// 	t.Error("client is nil")
	// }

	// if client != nil {
	// 	res, err := client.Run("uname -a")

	// 	if err != nil {
	// 		t.Error(err)
	// 	}

	// 	if res != nil {
	// 		t.Log(string(res))
	// 	}

	// }
}
