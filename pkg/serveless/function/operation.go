package function

import (
	"context"
	"fmt"
	"io"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/serveless/dockerregistry"
	"miniK8s/util/stringutil"
	"miniK8s/util/zip"
	"os"
	"path"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func (c *funcController) CreateFunction(f *apiObject.Function) error {
	// 【TODO】
	err := c.BuildFuncImage(f)

	if err != nil {
		return err
	}

	return nil
}

func (c *funcController) UpdateFunction(f *apiObject.Function) error {
	// 【TODO】
	err := c.DeleteFunction(f)

	if err != nil {
		return err
	}

	err = c.CreateFunction(f)

	if err != nil {
		return err
	}

	return nil
}

func (c *funcController) DeleteFunction(f *apiObject.Function) error {
	// 【TODO】
	return nil
}

func (c *funcController) BuildFuncImage(f *apiObject.Function) error {
	// 【TODO】
	// 在当前目录新建一个文件夹
	err := os.Mkdir(f.Metadata.UUID, 0777)

	if err != nil {
		return err
	}

	// 获取当前的路径
	curPath, err := os.Getwd()

	if err != nil {
		return err
	}

	// 将用户上传的文件放到这个文件夹中
	zip.ConvertBytesToZip(f.Spec.UserUploadFile, path.Join(curPath, f.Metadata.UUID, "userUploadFile.zip"))

	// 解压
	err = zip.DecompressZip(path.Join(curPath, f.Metadata.UUID, "userUploadFile.zip"), path.Join(curPath, f.Metadata.UUID))

	if err != nil {
		return err
	}

	// 将Dockerfile写入到这个文件夹中
	// 创建一个Dockerfile
	Dockerfile, err := os.Create(path.Join(curPath, f.Metadata.UUID, "Dockerfile"))

	if err != nil {
		return err
	}

	// 获取f.Spec.UserUploadFilePath的最后一个/后面的内容
	// 例如：/home/xxx/xxx/xxx/xxx.go
	// 获取的是xxx.go
	userUploadFolder := path.Base(f.Spec.UserUploadFilePath)
	localFolder := path.Join(f.Metadata.UUID, userUploadFolder, "*")
	containerFolder := "/app/"

	// 写入内容
	Dockerfile.WriteString("FROM musicminion/func-base:latest\n")
	Dockerfile.WriteString("COPY " + localFolder + " " + containerFolder + "\n")
	// EXPOSE 18080
	Dockerfile.WriteString("EXPOSE 18080\n")

	// 关闭文件
	Dockerfile.Close()

	// 构建镜像
	var cli *client.Client
	cli, _ = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	// 构建上下文
	contextDir := path.Join(curPath, f.Metadata.UUID)
	tarName := "context-" + stringutil.GenerateRandomStr(10) + ".tar"
	zip.CompressToTar(contextDir, "/tmp/"+tarName)
	fmt.Println(tarName)

	fmt.Println("开始构建上下文")
	tarBuf, err := os.Open("/tmp/" + tarName)

	if err != nil {
		return err
	}
	defer tarBuf.Close()

	fmt.Println("开始构建镜像")
	// 构建镜像
	resp, err := cli.ImageBuild(context.Background(), tarBuf, types.ImageBuildOptions{
		Tags:       []string{dockerregistry.Registry_Server_Prefix + "/func/" + f.Metadata.UUID + ":latest"},
		Dockerfile: path.Join(f.Metadata.UUID, "Dockerfile"),
		Remove:     true,
		Context:    tarBuf,
	})

	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("构建镜像成功")

	defer resp.Body.Close()
	_, err = io.Copy(os.Stdout, resp.Body)

	if err != nil {
		return err
	}

	// 推送镜像到镜像仓库
	err = dockerregistry.PushImageToRegistry(dockerregistry.Registry_Server_Prefix + "/func/" + f.Metadata.UUID + ":latest")

	if err != nil {
		return err
	}

	fmt.Println("推送镜像成功")

	// 删除文件夹
	// err = os.RemoveAll(path.Join(curPath, f.Metadata.UUID))

	return nil
}

// func createTarFromContext(contextDir string) (io.Reader, error) {
// 	tarBuf, err := os.CreateTemp("", "context"+stringutil.GenerateRandomStr(10)+".tar")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer tarBuf.Close()

// 	tw := tar.NewWriter(tarBuf)
// 	defer tw.Close()

// 	err = filepath.WalkDir(contextDir, func(path string, d os.DirEntry, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		relPath, err := filepath.Rel(contextDir, path)
// 		if err != nil {
// 			return err
// 		}

// 		info, err := d.Info()
// 		if err != nil {
// 			return err
// 		}

// 		header, err := tar.FileInfoHeader(info, relPath)
// 		if err != nil {
// 			return err
// 		}

// 		if err := tw.WriteHeader(header); err != nil {
// 			return err
// 		}

// 		if !d.IsDir() {
// 			file, err := os.Open(path)
// 			if err != nil {
// 				return err
// 			}
// 			defer file.Close()

// 			if _, err := io.Copy(tw, file); err != nil {
// 				return err
// 			}
// 		}

// 		return nil
// 	})

// 	if err != nil {
// 		return nil, err
// 	}

// 	_, err = tarBuf.Seek(0, 0)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return tarBuf, nil
// }
