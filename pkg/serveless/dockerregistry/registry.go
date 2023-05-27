package dockerregistry

import (
	"context"
	"encoding/json"
	"io"
	minik8stypes "miniK8s/pkg/minik8sTypes"
	"miniK8s/util/container"
	"net/http"
	"os"

	"github.com/docker/distribution/reference"
	rgclient "github.com/docker/distribution/registry/client"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var cli *client.Client

func init() {
	if cli == nil {
		cli, _ = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	}
	CheckRegistry()
}

// 初始化的时候检查，如果发现没有镜像管理中心的容器，则创建一个
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
		Env:          []string{"REGISTRY_STORAGE_DELETE_ENABLED=true"},
	})

	if err != nil {
		panic(err)
	}

	helper.ContainerManager.StartContainer(id)

	cli, _ = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

}

// 从镜像管理中心拉取镜像
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

// 把镜像推送到本地的miniK8s的镜像管理中心
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

// 从镜像管理中心删除镜像
func DeleteImageFromRegistry(imageName string) error {
	ctx := context.Background()
	ref, err := reference.Parse(imageName)
	if err != nil {
		return err
	}

	// 通过镜像名称获取镜像的仓库，然后通过仓库获取镜像的描述信息
	repo, err := rgclient.NewRepository(ref.(reference.Named), "http://"+Registry_Server_IP, http.DefaultTransport)

	if err != nil {
		return err
	}

	// 获取镜像的描述信息
	descpt, err := repo.Tags(ctx).Get(ctx, "latest")
	if err != nil {
		return err
	}

	manifestSvc, err := repo.Manifests(ctx, nil)

	if err != nil {
		return err
	}

	// 删除镜像
	err = manifestSvc.Delete(ctx, descpt.Digest)

	if err != nil {
		return err
	}

	return nil
}
