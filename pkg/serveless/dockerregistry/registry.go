package dockerregistry

import (
	"context"
	"encoding/json"
	"io"
	minik8stypes "miniK8s/pkg/minik8sTypes"
	"miniK8s/util/container"
	"net/http"
	"os"

	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	rgclient "github.com/docker/distribution/registry/client"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var cli *client.Client

// 创建一个docker image的镜像管理中心，用来管理用户上传的自定义的各种镜像
func CheckRegistry() {
	helper := container.NewHelperContainerManager()

	res, err := helper.ContainerManager.ListLocalContainers()

	if err != nil {
		panic(err)
	}

	// 检查镜像管理中心是否存在,如果不存在则创建
	for _, container := range res {
		if container.Names[0] == "/"+Registry_Server_Container_Name {
			if container.State == "running" {
				return
			} else {
				helper.ContainerManager.StartContainer(container.ID)
				return
			}
		}
	}

	// 发现不存在镜像管理中心的容器，创建一个

	// 检查镜像管理中心是否存在,如果不存在则创建
	helper.ImageManager.PullImageWithPolicy(Registry_Image_Name, minik8stypes.PullIfNotPresent)

	// 如果不存在则创建
	id, err := helper.ContainerManager.CreateHelperContainer(Registry_Server_Container_Name, &minik8stypes.ContainerConfig{
		Image: Registry_Image_Name,
		Tty:   false,
		PortBindings: map[nat.Port][]nat.PortBinding{nat.Port(Registry_Server_Port_Protocol): {
			{
				HostIP:   Registry_Server_BindIP,
				HostPort: Registry_Server_Port,
			},
		}},
		ExposedPorts: map[string]struct{}{Registry_Server_Port_Protocol: {}},
	})

	if err != nil {
		panic(err)
	}

	helper.ContainerManager.StartContainer(id)

	cli, _ = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

}

func PullImageFromRegistry(imageName string) error {
	authJson, err := json.Marshal(authInfo)

	if err != nil {
		return err
	}

	authJsonStr := string(authJson)

	info, err := cli.ImagePull(context.Background(), imageName, types.ImagePullOptions{
		RegistryAuth: authJsonStr,
		All:          false,
	})

	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, info)

	if err != nil {
		return err
	}

	return nil
}

func PushImageToRegistry(imageName string) error {
	authJson, err := json.Marshal(authInfo)

	if err != nil {
		return err
	}

	authJsonStr := string(authJson)

	info, err := cli.ImagePush(context.Background(), imageName, types.ImagePushOptions{
		RegistryAuth: authJsonStr,
		All:          false,
	})

	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, info)

	if err != nil {
		return err
	}

	return nil
}

func DeleteImageFromRegistry(imageName string) error {
	ctx := context.Background()

	repo, err := getImageRepository(imageName)

	if err != nil {
		return err
	}

	descpt, err := repo.Tags(ctx).Get(ctx, "latest")
	if err != nil {
		return err
	}

	manifestSvc, err := repo.Manifests(ctx, nil)

	if err != nil {
		return err
	}

	err = manifestSvc.Delete(ctx, descpt.Digest)

	if err != nil {
		return err
	}

	return nil
}

func ListImagesFromRegistry() error {

	return nil
}

func getImageRepository(imageName string) (distribution.Repository, error) {
	ref, err := reference.Parse(imageName)
	if err != nil {
		return nil, err
	}

	res, err := rgclient.NewRepository(ref.(reference.Named), "http://"+Registry_Server_IP, http.DefaultTransport)

	if err != nil {
		return nil, err
	}

	return res, nil
}
