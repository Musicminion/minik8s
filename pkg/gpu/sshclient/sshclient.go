package sshclient

import "github.com/melbahja/goph"

func NewSSHClient(username, password string) (*goph.Client, error) {
	cli, err := goph.NewUnknown(username, SSH_Login_URL, goph.Password(password))
	if err != nil {
		return nil, err
	}
	return cli, nil
}
