# 记录apiserver的开发流程

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

## TODO
###  apiObject
- 设计pod的数据结构
- 设计pod的handler 

###  解析yaml
- 通过go-yaml解析yaml文件
