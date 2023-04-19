package image

// 这是镜像管理相关的文件，主要是对镜像的相关操作
// 包括的函数有：
// 1. 拉取镜像 PullImage(imageStr string) (containerd.Image, error)
// 2. 获取镜像 GetImage(imageStr string) (containerd.Image, error)
// 3. 推送镜像 PushImage(imageStr string) error
// 4. 删除镜像 RemoveImage(imageStr string) error
// 5. 列出所有的镜像 ListAllImages() ([]containerd.Image, error)
// 后续如有需要再添加或者修改

import (
	"context"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/images"

	containerdClient "miniK8s/pkg/kubelet/containerdClient"
)

// 定义一个镜像管理者的结构体
type ImageManager struct {
}

// 实现相关的函数

// 拉取镜像，接受镜像的名词，返回错误
func (im *ImageManager) PullImage(imageStr string) (containerd.Image, error) {
	// client, err := containerd.New("/run/containerd/containerd.sock", containerd.WithDefaultNamespace("minik8s"))
	client, err := containerdClient.NewContainerdClient()
	// test, err :=

	if err != nil {
		return nil, err
	}

	defer client.Close()

	// 拉取镜像
	image, err := client.Pull(context.Background(), imageStr)
	if err != nil {
		return nil, err
	}

	// 返回镜像和错误信息
	return image, nil
}

// 获取镜像实例，镜像的实体
func (im *ImageManager) GetImage(imageStr string) (containerd.Image, error) {
	client, err := containerdClient.NewContainerdClient()
	if err != nil {
		return nil, err
	}

	defer client.Close()

	// 获取镜像
	image, err := client.GetImage(context.Background(), imageStr)
	if err != nil {
		return nil, err
	}
	return image, nil
}

// 推送镜像，没有错误返回nil
func (im *ImageManager) PushImage(imageStr string) error {
	client, err := containerdClient.NewContainerdClient()
	if err != nil {
		return err
	}

	defer client.Close()

	// 查找镜像
	image, err := client.GetImage(context.Background(), imageStr)

	if err != nil {
		return err
	}

	// 推送镜像
	err = client.Push(context.Background(), imageStr, image.Target())

	if err != nil {
		return err
	}

	return nil
}

// 删除镜像，没有错误返回nil
func (im *ImageManager) RemoveImage(imageStr string) error {
	client, err := containerdClient.NewContainerdClient()
	if err != nil {
		return err
	}

	defer client.Close()

	// 查找镜像
	image, err := client.GetImage(context.Background(), imageStr)

	if err != nil {
		return err
	}

	// 删除镜像
	err = client.ImageService().Delete(context.Background(), image.Name())
	if err != nil {
		return err
	}

	return nil
}

// 列出镜像，返回一个字符串切片
func (im *ImageManager) ListAllImages() ([]images.Image, error) {
	client, err := containerdClient.NewContainerdClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	result, err := client.ImageService().List(context.Background())
	if err != nil {
		return nil, err
	}
	return result, nil
}
