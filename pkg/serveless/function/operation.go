package function

import (
	"context"
	"errors"
	"fmt"
	"io"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/config"
	minik8stypes "miniK8s/pkg/minik8sTypes"
	"miniK8s/pkg/serveless/dockerregistry"
	netrequest "miniK8s/util/netRequest"
	"miniK8s/util/stringutil"
	"miniK8s/util/zip"
	"net/http"
	"os"
	"path"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// 新来一个函数，就创建一个函数
func (c *funcController) CreateFunction(f *apiObject.Function) error {
	// 【TODO】
	// 构建镜像
	err := c.BuildFuncImage(f)

	if err != nil {
		return err
	}

	// 创建副本
	err = c.CreateFuncReplica(f)

	if err != nil {
		return err
	}

	fmt.Println("create function success")
	return nil
}

// 对于函数的更新，就是删除旧的函数，然后创建新的函数
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

// 删除函数，就是删除函数的副本Replica
func (c *funcController) DeleteFunction(f *apiObject.Function) error {
	// 【TODO】
	err := c.DeleteFuncReplica(f)

	if err != nil {
		return err
	}

	return nil
}

// 构建函数的镜像
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
	err = os.RemoveAll(path.Join(curPath, f.Metadata.UUID))

	if err != nil {
		return err
	}

	return nil
}

func (c *funcController) CreateFuncReplica(f *apiObject.Function) error {
	// 【TODO】
	replica := &apiObject.ReplicaSet{
		Basic: apiObject.Basic{
			Kind:       apiObject.ReplicaSetKind,
			APIVersion: serverconfig.APIVersion,
			Metadata: apiObject.Metadata{
				Name:      f.Metadata.Name,
				Namespace: f.Metadata.Namespace,
				Labels: map[string]string{
					minik8stypes.Replica_Func_Name:      f.Metadata.Name,
					minik8stypes.Replica_Func_Uuid:      f.Metadata.UUID,
					minik8stypes.Replica_Func_Namespace: f.Metadata.Namespace,
				},
			},
		},
		Spec: apiObject.ReplicaSetSpec{
			Replicas: 2,
			Selector: apiObject.ReplicaSetSelector{
				MatchLabels: map[string]string{
					minik8stypes.Pod_Func_Name:      f.Metadata.Name,
					minik8stypes.Pod_Func_Uuid:      f.Metadata.UUID,
					minik8stypes.Pod_Func_Namespace: f.Metadata.Namespace,
				},
			},
			Template: apiObject.PodTemplate{
				Metadata: apiObject.Metadata{
					Name:      f.Metadata.Name + "-" +stringutil.GenerateRandomStr(10),
					Namespace: f.Metadata.Namespace,
					Labels: map[string]string{
						minik8stypes.Pod_Func_Name:      f.Metadata.Name,
						minik8stypes.Pod_Func_Uuid:      f.Metadata.UUID,
						minik8stypes.Pod_Func_Namespace: f.Metadata.Namespace,
					},
				},
				Spec: apiObject.PodSpec{
					Containers: []apiObject.Container{
						{
							Name:  f.Metadata.Name + "-container-" + stringutil.GenerateRandomStr(8),
							Image: dockerregistry.Registry_Server_Prefix + "/func/" + f.Metadata.UUID + ":latest",
							Ports: []apiObject.ContainerPort{
								{
									ContainerPort: "18080",
								},
							},
						},
					},
				},
			},
		},
	}

	url := config.GetAPIServerURLPrefix() + config.ReplicaSetsURL
	url = stringutil.Replace(url, config.URL_PARAM_NAMESPACE_PART, f.Metadata.Namespace)

	code, _, err := netrequest.PostRequestByTarget(url, replica)

	if err != nil {
		return err
	}

	if code != http.StatusCreated {
		return errors.New("创建副本集失败, Code Not 201")
	}

	return nil
}

func (c *funcController) DeleteFuncReplica(f *apiObject.Function) error {
	// 【TODO】
	url := config.GetAPIServerURLPrefix() + config.ReplicaSetSpecURL
	url = stringutil.Replace(url, config.URL_PARAM_NAMESPACE_PART, f.Metadata.Namespace)
	url = stringutil.Replace(url, config.URL_PARAM_NAME_PART, f.Metadata.Name)

	code, err := netrequest.DelRequest(url)

	if err != nil {
		return err
	}

	if code != http.StatusNoContent {
		return errors.New("删除副本集失败, Code Not 204")
	}

	return nil
}
