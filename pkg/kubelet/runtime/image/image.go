package image

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	dockerclient "miniK8s/pkg/kubelet/dockerClient"
	imageTypes "miniK8s/pkg/minik8sTypes"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

type ImageManager struct {
}

// PullImageWithPolicy从镜像仓库拉取镜像
// 如果镜像已经存在，那么根据策略决定是否重新拉取
// 如果镜像不存在，那么就拉取镜像
// 如果镜像不存在，且策略是ImagePullPolicyNever，那么就返回错误
func (im *ImageManager) PullImageWithPolicy(imageRef string, policy imageTypes.ImagePullPolicy) (string, error) {
	ctx := context.Background()
	client, err := dockerclient.NewDockerClient()
	if err != nil {
		return "", err
	}

	defer client.Close()

	switch policy {
	// 无论如何都要拉取镜像
	case imageTypes.PullAlways:
		// 拉取镜像
		image, err := client.ImagePull(ctx, imageRef, types.ImagePullOptions{})
		// println(imageRef)
		if err != nil {
			return "", err
		}

		file, err := os.Create(os.DevNull)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		io.Copy(file, image)
		defer image.Close()

		imageIDs, err := im.findLocalImageIDsByImageRef(imageRef)

		if err != nil {
			return "", err
		}

		if len(imageIDs) != 1 {
			return "", errors.New("imageID not found or more than one imageID")
		}

		return imageIDs[0], nil

	// 如果镜像不存在，那么就拉取镜像
	case imageTypes.PullIfNotPresent, "":
		// 在本地查找镜像
		imageIDs, err := im.findLocalImageIDsByImageRef(imageRef)
		if err != nil {
			return "", err
		}
		if len(imageIDs) == 0 {
			// 拉取镜像
			image, err := client.ImagePull(ctx, imageRef, types.ImagePullOptions{})

			if err != nil {
				return "", err
			}
			io.Copy(os.Stdout, image)

			defer image.Close()

			imageIDs, err := im.findLocalImageIDsByImageRef(imageRef)

			if err != nil {
				return "", err
			}

			if len(imageIDs) != 1 {
				fmt.Println(imageIDs)
				return "", errors.New("imageID not found or more than one imageID")
			}

			return imageIDs[0], nil
		} else {
			return imageIDs[0], nil
		}
	case imageTypes.PullNever:
		// 在本地查找镜像
		imageIDs, err := im.findLocalImageIDsByImageRef(imageRef)
		if err != nil {
			return "", err
		}
		if len(imageIDs) == 0 {
			// 返回一个空的imageID和错误
			return "", errors.New("image not found")
		}
		return imageIDs[0], nil
	}

	// 创建一个错误返回
	err = errors.New("policy not found or not supported")
	return "", err
}

// 删除镜像,没有错误就返回nil
func (im *ImageManager) RemoveImage(imageRef string) error {
	ctx := context.Background()
	client, err := dockerclient.NewDockerClient()
	if err != nil {
		return err
	}

	defer client.Close()

	// 获取imageID
	imageIDs, err := im.findLocalImageIDsByImageRef(imageRef)

	if err != nil {
		return err
	}

	// 找到的镜像数量不为1，那么就返回错误
	if len(imageIDs) != 1 {
		return errors.New("image not found or found more than one")
	}

	// 删除镜像
	_, err = client.ImageRemove(ctx, imageIDs[0], types.ImageRemoveOptions{})

	if err != nil {
		return err
	}

	return nil
}

// 通过ImageRef查找本地镜像,返回镜像ID的切片
func (im *ImageManager) findLocalImageIDsByImageRef(imageRef string) ([]string, error) {
	ctx := context.Background()
	client, err := dockerclient.NewDockerClient()
	if err != nil {
		// 返回一个空的切片和错误
		return []string{}, err
	}

	defer client.Close()
	// 创建一个过滤器
	filterArgs := filters.NewArgs()
	filterArgs.Add("reference", parseImageRef(imageRef))
	// 查找镜像

	images, err := client.ImageList(ctx, types.ImageListOptions{
		Filters: filterArgs,
	})

	// 创建一个空的切片
	imageIDs := []string{}

	// 遍历输出的镜像
	for _, image := range images {
		imageIDs = append(imageIDs, image.ID)
	}

	if err != nil {
		return []string{}, err
	}

	return imageIDs, nil
}

// 工具函数
// 解析镜像"docker.io/library/busybox:latest", 返回 "busybox:latest"
func parseImageRef(imageRef string) (result string) {
	// 检查imageRef是否是docker.io开头的
	if !strings.HasPrefix(imageRef, "docker.io/") {
		result = imageRef
		return result
	}

	// 通过/分割字符串
	split := strings.Split(imageRef, "/")
	// 如果分割后的长度大于1，那么就取最后一个
	if len(split) > 1 {
		result = split[len(split)-1]
	} else {
		result = imageRef
	}
	return result
}
