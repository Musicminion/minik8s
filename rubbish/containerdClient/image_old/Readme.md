## imgae组件说明

这是镜像管理相关的文件，主要是对镜像的相关操作
包括的函数有：
1. 拉取镜像 PullImage(imageStr string) (containerd.Image, error)
2. 获取镜像 GetImage(imageStr string) (containerd.Image, error)
3. 推送镜像 PushImage(imageStr string) error
4. 删除镜像 RemoveImage(imageStr string) error
5. 列出所有的镜像 ListAllImages() ([]containerd.Image, error)

后续如有需要再添加或者修改，目前经过测试全部通过。

### 测试逻辑
如果修改代码检查发现测试的时候报错，可能可以参考测试的逻辑
1. 测试前，首先会删除所有的已经存在的 `miniK8s` 名字空间下面的镜像！注意！测试需谨慎
2. 依次执行测试TestPullImage，拉取镜像，拉取的镜像包含下面的几个内容

```go
var TestImageURLs = []string{
	"docker.io/library/busybox:latest",
	"docker.io/library/hello-world:latest",
	"docker.io/library/nginx:latest",
	"docker.io/library/redis:latest",
}
```

3. 然后测试TestGetImage，相当于逐个获取镜像的实体，并检查名字是否和上面的数组匹配
4. 接下来测试TestListAllImages，一次性获取所有的镜像
5. 最后测试TestRemoveImage，相当于逐个的删除所有镜像
6. **目前没有测试Push镜像，因为目前项目里面没有要用到这个函数的，届时可能会删除接口！**

### 测试逻辑Todo
1. 后续可能需要根据条件筛选镜像，届时根据开发的情况而定
2. Push可能需要权限认证