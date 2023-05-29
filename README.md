# Minik8s

<img src="https://wakatime.com/badge/user/485d951d-d928-4160-b75c-855525f5ae42/project/334b3ff9-9175-48b2-9f54-cc38a9244d7d.svg" alt=""/> <img src="https://img.shields.io/badge/go-1.20-blue" alt=""/>

>  2023年《SE3356 云操作系统设计与实践》课程第一小组项目。

小组成员如下：

- 董云鹏 517021910011 [@dongyunpeng-sjtu](https://github.com/dongyunpeng-sjtu)
- 冯逸飞 520030910021 [@every-breaking-wave](https://github.com/every-breaking-wave)
- 张子谦 520111910121 [@Musicminion](https://github.com/Musicminion)

项目仓库的地址：
- Github: https://github.com/Musicminion/minik8s/
- Gitee: https://gitee.com/Musicminion/miniK8s

项目的CI/CD主要在Github上面运行，所以如有需要查看，请移步到Github查看。

## 架构
### 使用到的开源库

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

### 架构

**开发语言**：Golang。我们项目主要采用go语言(版本1.20)进行开发。之所以选择go语言，因为docker、k8s也是基于go开发的，并且docker提供了go相关的sdk，让我们轻松就能将项目接入，实现通过go语言来操作底层的容器运行、获取运行状态等信息。另外go语言本身具有很好的容错能力，通过强制检查返回值`err`的方式，也让错误报错更加友好。

**项目架构**：我们的项目架构学习了K8s的架构，同时又适应需求做了一定的微调。整体主要是由：控制平面和Worker节点的两大类组成。运行在控制平面的组件主要有下面的几个：

- API Server：提供一系列Restful的接口，例如对于API对象的增删改查接口，供其他组件使用
- Controller：包括DNS Controller、HPA Controller、Replica Controller、JobController，主要是对于一些抽象级别的API对象的管理，状态的维护。
- Scheduler：负责从所有的可以使用的节点中，根据一定的调度策略，当收到Pod调度请求时，返回合适的节点
- Serveless：单独运行的一个服务器，负责维护Serveless的函数相关对象的管理，同时负责转发用户的请求到合适的Pod来处理
- RabbitMQ：作为消息队列，集群内部的消息的通讯工具

运行在WorkerNode上面的主要有下面的几个组件

- kubeproxy：负责DNS、Iptable的修改，维护Service的状态等
- Kubelet：维护Pod的底层创建，Pod生命周期的管理，Pod异常的重启/重建等
- Redis：作为本地的缓存Cache，哪怕API-Server完全崩溃，因为有本地的Redis，机器重新启动之后，Kubelet也能够恢复之前容器的状态

<img width="1071" alt="截屏2023-05-29 00 09 32" src="https://github.com/Musicminion/minik8s/assets/84625273/9741e107-dc8d-4bfb-b71f-819c6956c1b5">


**项目分支**：我们的开发采用多分支进行。每一个功能点严格对应一个Feature分支，所有的推送都会经过`go test`的测试检验。并可以在[这里](https://github.com/Musicminion/minik8s/actions)查看详细的情况。

项目一共包含主要分支包括
- Master分支：项目的发行分支，**只有通过了测试**,才能通过PR合并到Master分支。
- Development分支：开发分支，用于合并多个Feature的中间分支，
- Feature/* 分支：功能特定分支，包含相关功能的开发分支

如下图所示，是我们开发时候的Pr合并的情况。所有的Pr都带有相关的Label，便于合并的时候审查。考虑到后期的合并比较频繁，我们几乎都是每天都需要合并最新的工作代码到Development分支，然后运行单元测试。测试通过之后再合并到Master分支。

![image](https://github.com/Musicminion/minik8s/assets/84625273/e656077e-adb1-4030-ba45-194a791c6d60)


**CI/CD介绍**：CI/CD作为我们软件质量的重要保证之一。我们通过Git Action添加了自己的Runner，并编写了项目的测试脚本来实现CI/CD。保证每次运行前环境全部初始化。
- 所有的日常代码的推送都会被发送到我们自己的服务器，运行单元测试，并直接显示在单次推送的结果后方
- 当发起Pr时，自动会再一次运行单元测试，测试通过之后才可以合并
- 运行单元测试通过之后，构建可执行文件，发布到机器的bin目录下
- 以上2,3条通过之后，对于合并到Master的情况，会构建docker相关的镜像(例如GPU Job的docker镜像、Function的基础镜像)推送到dockerhub

**软件测试介绍**：go语言自身支持测试框架。并且鼓励把项目文件和测试文件放在同一个文件夹下面。例如某一个项目的文件是file.go,那么测试的文件的名字就是file_test.go。最终要运行整个项目测试的时候，只需要在项目的根目录运行 `go test ./...` 即可完成整个项目的测试。测试会输出详细的测试通过率，非常方便。

**功能开发流程**：
- 我们的软件开发基于迭代开发、敏捷开发。小组成员每天晚上在软件学院大楼实验室集中进行开发新功能，减少沟通障碍，做到有问题及时解决、沟通，有困难相互请教，这也大大的提高了我们小组的效率。截止15周周末，我们已经完成了所有的功能的开发。基本符合预期进度。
- 对于新功能开发，我们采用"动态分配"方法，根据进度灵活分配成员的任务。项目框架搭建好之后，基本上在任何时间点小组同时在开发两个或者两个以上的需求。一人开发完成之后，交给另外一个组员完成代码的审查和测试，测试通过之后合并到Master
- 功能开发的过程主要是：简要的需求分析->设计API对象->设计API-Server的接口->设计Etcd存储情况->编写该需求的运行逻辑代码->编写Kubectl相关代码->最终测试
- 具体如下图所示，在整个开发的流程中，我们基本都是在重复下面的流程图。

<img width="865" alt="截屏2023-05-29 00 14 54" src="https://github.com/Musicminion/minik8s/assets/84625273/13a63fac-58de-4747-905c-e932f0a830f9">

**开发简介**：
- 项目代码体量大约2w行代码，开发周期大约1.5月
- 完成要求里面的全部功能

### 组件详解

#### API Server

**API-Server**：API Server是minik8s控制平面的核心。主要负责和etcd存储打交道，并提供一些核心的APIObject的API，供其他组件使用。在设计API Server的API时候，我们主要考虑了两个特性，一个是状态(Status)和期望(Spec)分离的情况，另外一个是Etcd的路径和API分离。

如果没有分离，我们考虑下面的情景：当一个Kubelet想要更新某一个Pod的状态的时候，试图通过Post或者Put请求写入一个完整的Pod对象，在此之前,假如用户刚刚通过`kubectl apply`更新了一个Pod的信息，如果按照上面我所叙述的时间线，就会出现用户的apply的更新的Pod被覆盖了。同样的道理，如果用户删除了一个Pod，按照上面的设计，Kubelet在回传的时候写入了一个完整的Pod，相当于没有做任何的删除。

虽然说上面的例子是因为期望和状态没有分离，但是本质是kubelet的权限太大，能够写入一个完整的Pod。所以为了解决这种问题，我们对于一个对象，往往设计了更新对象接口(更新整个对象)，更新对象的状态接口(仅仅更新Status，如果找不到对象那么就不更新)

第二个设计时分开了API的格式和Etcd存储API对象的路径。Etcd存储API对象的路径都是诸如`registry/pods/<namespace>/<name>`，而API的格式大多都是`/api/v1/pods/namespaces/:namespace/name/:name`,可以看到两者的差别还是比较明显的，这是因为API版本看发生动态变化(在实际的k8s中也是这样)，但是存储的路径保证兼容原来的。所以在我们的minik8s中，我们同样借鉴了这样的思路。

更多有关API-Server的内容以及详细的API文档，请移步到`/pkg/apiserver`下的Readme查看。

#### kubelet架构
**kubelet**：Kubelet是和容器底层运行打交道的组件，确保每一个Pod能够在该节点正常运行。目前kubelet架构设计如下(我们参考了k8s的反馈路径设计并做了一定的微调，以适应项目)

- Kubelet主要由：StatusManager、RunTimeManager、PlegManager、WorkerManager几个核心组件和Pleg、MsgChan的通道组成。
- RunTimeManager和底层的Docker交互，用于创建容器、获取容器运行的状态、管理镜像等操作
- WorkerManager用于管理Worker，我们的策略是每一个Pod分配一个Worker，然后由WorkerManager进行统一的调度和分配。每一个Worker有他自己的通道，当收到Pod的创建或者删除任务的时候，就会执行相关的操作
- PlegManager用来产生Pleg(Pod LifeCycle Event)，发送到PlegChan。PlegManager会调用StatusManager，比较缓存里面的Pod的情况和底层运行的Pod的情况，产生相关的事件。
- ListWatcher会监听属于每个Node的消息队列，当收到创建/删除Pod的请求的时候，也会发送给相关的WorkerManager
- 也就是说创建Pod会有消息队列/StatusManager检测到和远端不一致这样两种路径，前者的效率更高，后者用于维护长期的稳定。两者协同保证Pod的正确运行

<img width="1226" alt="截屏2023-05-29 08 48 24" src="https://github.com/Musicminion/minik8s/assets/84625273/352821f7-a9a3-4151-8fac-89a5e753184a">

<img width="320"  align='right'  alt="截屏2023-05-29 08 50 13" src="https://github.com/Musicminion/minik8s/assets/84625273/92b2e789-8f23-4314-a388-be011e08af21">

具体来说，各个组件之间的行为和关系如下图详细所示。
- Runtime Manager会负责收集底层正在运行的所有的容器的信息，并把容器的信息组装为Pod的状态信息。同时收集当前机器的CPU/内存状态，把相关的信息回传到API Server，及时更新。
- Status Manager还会定期的从API Server拉取当前节点上所有的Pod，以便于比较和对齐，产生相关的容器生命周期事件(Pleg)，
- Status Manager对于所有更新获取到的Pod，都会写入Redis的本地缓存，以便于API-Server完全崩溃和Kubelet完全崩溃重启的时候，Kubelet有Pod的期望信息，能够作为对齐目标
- 当出现Pod不一致的时候，以远端的API-Server的数据为主，并清除掉不必要的Pod。如下图所示，会清空不必要的Pod，并创建本地没有的Pod，实现和远端数据的对齐。

<img width="1017" alt="截屏2023-05-29 09 00 41" src="https://github.com/Musicminion/minik8s/assets/84625273/add8f63b-b90e-46d7-9971-2d78e7ef1c59">

#### controller架构
minik8s需要controller对一些抽象的对象实施管理。Controller是运行在控制平面的一个组件，具体包括DNS Controller、HPA Controller、Job Controller、Replica Controller。之所以需要Controller来对于这些API对象进行管理，是因为这些对象都是比较高度抽象的对象，需要维护已有的基础对象和他们之间的关系，或者需要对整个系统运行状态分析之后再才能做出决策。具体的逻辑如下：
- Replica Controller：维护Replica的数量和期望的数量一直，如果出现数量不一致，当通过标签匹配到的Pod数量较多的时候，会随机的杀掉若干Pod，直到数量和期望一致；当通过标签匹配到的Pod数量偏少的时候，会根据template创建相关的Pod
- Job Controller：维护GPU Job的运行，当一个新的任务出现的时候，会被GPU JobController捕捉到（因为这个任务没有被执行，状态是空的），然后Controller会创建一个新的Pod，让该Pod执行相关的GPU任务。
- HPA Controller：分析HPA对应的Pod的CPU/Mem的比例，如果出现了异常，就会出发扩容或者缩容。所有的扩容、所有都是以一个Pod为单位进行的，并且默认的扩容/缩容的速度是15s/单位Pod。如果用户自己指定了扩缩容的速度，那么遵循用户的规则。
- DNS Controller：TODO

#### Kubectl

Kubectl作为minik8s的命令行管理工具，命令的设计基本参考kubernates。我们使用了Cobra的命令行解析工具，大大提高了命令解析的效率。

<img width="300" alt="截屏2023-05-29 09 00 41" src="https://github.com/Musicminion/minik8s/assets/84625273/7b7fe250-0999-4c68-ae3a-e1f481028769">

支持的命令如下所示：
- `Kubectl apply ./path/to/your.yaml` 创建一个API对象，会自动识别文件中对象的Kind，发送给对应的接口
- `Kubectl delete ./path/to/your.yaml` 根据文件删除一个API对象，会自动识别文件中对象的name和namespace，发送给对应的接口(删除不会校验其他字段是否完全一致)
- `kubectl get [APIObject] [Namespace/Name]` 获取一个API对象的状态(显示经过简化的信息，要查看详细的结果，请使用Describe命令)
- `kubectl describe [APIObject] [Namespace/Name]` 获取一个API对象的详细的json信息(显示完整的经过优化的json字段)
- `kubectl execute [namespace]/[name] [parameters]` 出发一个Serveless的函数，并传递相关的参数

#### Scheduler
Scheduler是运行在控制平面负责调度Pod到具体Node的组件。Scheduler和API-Server通过RabbitMQ消息队列实现通讯。当有Pod创建的请求的时候，API-Server会给Scheduler发送调度请求，Scheduler会主动拉取所有Node，根据最新的Node Status来安排调度。

目前我们的Scheduler支持多种调度策略：
- RoundRobin：轮询调度策略
- Random：随机调度策略
- LeastPod：选择Pod数量最少的节点
- LeastCpu：选择CPU使用率最低的作为调度目标
- LeastMem：选择Mem使用率最低的作为调度的目标

这些调度策略可以通过启动时候的参数传递，以便于Scheduler知道以哪一种调度策略运行。


#### Kuberproxy

### 需求实现详解

#### Pod抽象

Pod的演示视频请参考：

Pod是k8s(minik8s)调度的最小单位。用户可以通过`Kubectl apply Podfile.yaml`的声明式的方法创建一个Pod。当用户执行该命令后，Kubectl会将创建Pod的请求发送给API-Server。API-Server检查新创建的Pod在格式、字段是否存在问题，如果没有异常，就会写入Etcd，并给Scheduler发送消息。

Scheduler完成调度之后，会通过消息队列通知API-Server，API-Server收到调度结果，将对应的Pod的nodename字段写入调度结果，然后保存回Etcd。然后主动给相关的Kubelet发送Pod的创建请求。

之前已经介绍到Kubelet创建Pod可以有两条途径，一条是长期拉取自己节点所有的Pod，另外一条途径是收到消息队列的创建请求之后主动创建。这两条途径**不会冲突**，因为在WorkManager底层是每一个Pod对应一个Worker，一旦收到了创建请求，再次收到创建请求的时候就会被忽略。Kubelet收到创建Pod请求之后，会把Pod的配置信息写入到本地的Redis里面，这样即使是API-Server崩溃，Kubelet出现重启，也能够保证Pod的信息可以读取到。

Pod创建之后，Kubelet会不断监视Pod的运行状态，并将状态更新写回到API-Server(通过Pod的Status的接口)
