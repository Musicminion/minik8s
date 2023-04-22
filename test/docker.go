package main

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	imageName := "docker.io/library/ubuntu:latest"

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	defer out.Close()
	io.Copy(&os.NewFile(), out)

	// resp, err := cli.ContainerCreate(ctx, &container.Config{
	// 	Image: imageName,
	// 	Cmd:   []string{"/bin/bash", "-c", "while true; do echo hello world; sleep 1; done"},
	// }, nil, nil, nil, "")
	// if err != nil {
	// 	panic(err)
	// }

	// if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
	// 	panic(err)
	// }

	// fmt.Println(resp.ID)
}
