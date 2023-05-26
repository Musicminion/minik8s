package function

import (
	"miniK8s/pkg/apiObject"
	"miniK8s/util/zip"
	"os"
	"path"
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

	return nil
}
