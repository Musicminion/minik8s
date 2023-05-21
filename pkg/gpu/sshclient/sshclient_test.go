package sshclient

import (
	"os"
	"testing"
)

const ifTest = false

func TestEnv(t *testing.T) {
	t.Log("GPU_SSH_USERNAME: ", os.Getenv("GPU_SSH_USERNAME"))
	t.Log("GPU_SSH_PASSWORD: ", os.Getenv("GPU_SSH_PASSWORD"))
}

func TestNewSSHClient(t *testing.T) {
	if ifTest == false {
		return
	}
	client, err := NewSSHClient(os.Getenv("GPU_SSH_USERNAME"), os.Getenv("GPU_SSH_PASSWORD"))

	if err != nil {
		t.Error(err)
	}

	if client == nil {
		t.Error("client is nil")
	}

	if client != nil {
		res, err := client.RunCmd("uname -a")

		if err != nil {
			t.Error(err)
		}

		if res != "" {
			t.Log(string(res))
		}

	}
}

func TestFileBasic(t *testing.T) {
	if ifTest == false {
		return
	}

	client, err := NewSSHClient(os.Getenv("GPU_SSH_USERNAME"), os.Getenv("GPU_SSH_PASSWORD"))

	if err != nil {
		t.Error(err)
	}

	if client == nil {
		t.Error("client is nil")
	}

	if client != nil {
		res, err := client.MakeDirectory("minik8s")
		if err != nil {
			t.Error(err)
		}
		t.Log(string(res))

		res, err = client.ChangeDirectory("minik8s")
		if err != nil {
			t.Error(err)
		}
		t.Log(string(res))

		res, err = client.WriteFile("test.txt", "hello world zzq")
		if err != nil {
			t.Error(err)
		}
		t.Log(string(res))
		res, err = client.AppendFile("test.txt", "hello world !!!")
		if err != nil {
			t.Error(err)
		}
		t.Log(string(res))
		res, err = client.ReadFile("test.txt")
		if err != nil {
			t.Error(err)
		}
		t.Log(string(res))
		res, err = client.RemoveFile("test.txt")
		if err != nil {
			t.Error(err)
		}
		t.Log(string(res))
		res, err = client.ChangeDirectory("..")
		if err != nil {
			t.Error(err)
		}
		t.Log(string(res))
		res, err = client.RemoveDirectory("minik8s")
		if err != nil {
			t.Error(err)
		}
		t.Log(string(res))

	}
}
