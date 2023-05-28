## Minik8s

<img src="https://wakatime.com/badge/user/485d951d-d928-4160-b75c-855525f5ae42/project/334b3ff9-9175-48b2-9f54-cc38a9244d7d.svg" alt=""/> <img src="https://img.shields.io/badge/go-1.20-blue" alt=""/>

>  2023年《SE3356 云操作系统设计与实践》课程第一小组项目。

小组成员如下：

- 董云鹏 517021910011
- 冯逸飞 520030910021
- 张子谦 520111910121

项目仓库的地址：
- Github: https://github.com/Musicminion/minik8s/
- Gitee: https://gitee.com/Musicminion/miniK8s

项目的CI/CD主要在Github上面运行，所以如有需要查看，请移步到Github查看。

### 架构
#### 使用到的开源库

- [github.com/docker/docker](https://github.com/moby/moby) 底层容器运行时的操作
- [github.com/pallets/flask](https://github.com/pallets/flask) Serveless容器内的运行的程序
- [github.com/gin-gonic/gin](https://github.com/gin-gonic/gin) APIServer框架
- [github.com/fatih/color](https://github.com/fatih/color) minik8s的分级日志系统
- [github.com/klauspost/pgzip](https://github.com/klauspost/pgzip) 用户文件的zip压缩
- [github.com/melbahja/goph](https://github.com/melbahja/goph) GPU Job的SSH的客户端
- [github.com/mholt/archiver](https://github.com/mholt/archiver) Docker镜像打包时用到的tar压缩
- [go.etcd.io/etcd/client/v3](https://github.com/etcd-io/etcd) 和Etcd存储交互操作的客户端
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) go的yaml文件解析
- [gotest.tools/v3](https://github.com/gotestyourself/gotest.tools) 项目测试框架
- [docker/login-action](https://github.com/docker/login-action) CICD自动推送镜像到dockerHub
- [docker/setup-qemu-action ](https://github.com/docker/setup-qemu-action)CICD交叉编译平台
- [github.com/google/uuid](https://github.com/google/uuid) API对象UUID的生成
- [github.com/spf13/cobra](https://github.com/spf13/cobra) Kubectl的命令行工具
- [github.com/jedib0t/go-pretty/table](https://github.com/jedib0t/go-pretty/table) Kubectl美化输出

#### 架构

**开发语言**：Golang。我们项目主要采用go语言(版本1.20)进行开发。之所以选择go语言，因为docker、k8s也是基于go开发的，并且docker提供了go相关的sdk，让我们轻松就能将项目接入，实现通过go语言来操作底层的容器运行、获取运行状态等信息。

**项目架构**：我们的项目架构学习了K8s的架构，同时又适应需求做了一定的微调。主要是由：控制平面和Worker节点的两大类组成。运行在控制平面的组件主要有下面的几个：

- API Server：提供一系列Restful的接口，例如对于API对象的增删改查接口
- Controller：包括DNS Controller、HPA Controller、Replica Controller、JobController，主要是对于一些抽象级别的API对象的管理，状态的维护。
- Scheduler：负责从所有的可以使用的节点中，根据一定的调度策略，当收到Pod调度请求时，返回合适的节点
- Serveless：单独运行的一个服务器，负责维护Serveless的函数相关对象的管理，同时负责转发用户的请求到合适的Pod来处理
- RabbitMQ：作为消息队列，集群内部的消息的通讯工具

运行在WorkerNode上面的主要有下面的几个组件

- kubeproxy：负责DNS、Iptable的修改，维护Service的状态等
- Kubelet：维护Pod的底层创建，Pod生命周期的管理，Pod异常的重启/重建等
- Redis：作为本地的缓存，哪怕API-Server完全崩溃，因为有本地的Redis，机器重新启动之后，Kubelet也能够恢复之前容器的状态

**项目分支**：我们的开发采用多分支进行。每一个功能点严格对应一个Feature分支，所有的推送都会经过`go test`的测试检验。并可以在[这里]（https://github.com/Musicminion/minik8s/actions)查看详细的情况。

项目一共包含主要分支包括
- Master分支：项目的发行分支，**只有通过了测试**,才能通过PR合并到Master分支。
- Development分支：开发分支，用于合并多个Feature的中间分支，
- Feature/* 分支：功能特定分支，包含相关功能的开发分支

如下图所示，是我们开发时候的Pr合并的情况。所有的Pr都带有相关的Label，便于合并的时候审查。考虑到后期的合并比较频繁，我们几乎都是每天都需要合并最新的工作代码到Development分支，然后运行单元测试。测试通过之后再合并到Master分支。

![image](https://github.com/Musicminion/minik8s/assets/84625273/e656077e-adb1-4030-ba45-194a791c6d60)


**CI/CD介绍**：CI/CD作为我们软件质量的重要保证之一。我们通过Git Action添加了自己的Runner，并编写了项目的测试脚本来实现CI/CD。
- 所有的日常代码的推送都会被发送到我们自己的服务器，运行单元测试，并直接显示在单次推送的结果后方
- 当发起Pr时，自动会再一次运行单元测试，测试通过之后才可以合并
- 运行单元测试通过之后，构建可执行文件，发布到机器的bin目录下
- 以上两条通过之后，构建docker相关的镜像(例如GPU Job的docker镜像、Function的基础镜像)推送到dockerhub




### 记录apiserver的开发流程

### kubelet架构
目前kubelet架构设计如下

<img width="1021" alt="截屏2023-05-03 23 32 05" src="https://user-images.githubusercontent.com/84625273/235964773-d77faaec-c39d-4778-859f-1387bfdf24d3.png">

### 已完成

### 对etcd的接口封装
- put
- get
- del
- watch
- prefix_related

### naive apiserver
- 支持基本的http请求
- 配置了server的一些默认配置
- 设置了用于测试的handle

## TODOz
###  apiObject
- 设计pod的数据结构
- 设计pod的handler 

###  解析yaml
- 通过go-yaml解析yaml文件
