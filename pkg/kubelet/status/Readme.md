### 对于StatusManager的问题
1. StatusManger是干什么的
回答：StatusManager是Kubelet里面和API Server直接进行Http通信的组件，核心交互的功能有下面的三个
- 发布自己的节点的运行状态（简称：节点心跳包）
- 发布运行在自己节点的Pod（告知API Server自己上Pod的运行状态）：只会更新状态，不会涉及到创建和删除。如果更新一个不存在的Pod状态，服务端会返回错误（为什么这样设计，否则删除容器之后，中间有一段时间，Node还在发送Pod的状态，所以必须把状态和Pod的情况分开）
- 拉取API Server中，自己节点上最新的Pod的信息，具体的更新策略是：

![image](https://github.com/Musicminion/minik8s/assets/84625273/04a20e9d-3703-4874-914d-9e123a2f76d5)

- 远端有，本地没有，那就创建
- 远端没有，本地有，那就删除杀死Pod
- 远端有，本地有，那就选择时间最新的缓存
