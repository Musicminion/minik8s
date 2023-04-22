package dockerclient

import "github.com/docker/docker/client"

/*
 * 创建一个docker client
 */
func NewDockerClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return cli, nil
}
