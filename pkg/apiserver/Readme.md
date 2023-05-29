# minik8s APIServer

API Server是minik8s控制平面的核心。主要负责和ETCD存贮打交道，并提供一些核心的APIObject的API，供其他组件使用。在设计API Server的时候，我们主要考虑了两个特性，一个是状态和期望分离的情况，另外一个是Etcd的路径和API分离。

因为如果没有分离，我们考虑下面的情景：当一个Kubelet想要更新某一个Pod的状态的时候，试图Post或者Put请求写入一个完整的Pod对象，在此之前假如用户刚刚通过`kubectl apply`更新了一个Pod的信息，如果按照上面我所叙述的时间线，就会出现用户的apply的更新的Pod被覆盖了。同样的道理，如果用户删除了一个Pod，按照上面的设计，Kubelet在回传的时候写入了一个完整的Pod，相当于没有做任何的删除。

虽然说上面的例子是因为期望和状态没有分离，但是本质是kubelet的权限太大，能够写入一个完整的Pod。所以为了解决这种问题，我们对于一个对象，往往设计了更新对象接口(更新整个对象)，更新对象的状态接口(仅仅更新Status，如果找不到对象那么久不更新。)

Etcd存储API对象的路径都是`registry/pods/<namespace>/<name>`，而API的格式大多都是`/api/v1/pods/namespaces/:namespace/name/:name`,可以看到两者的差别还是比较明显的，这是因为API版本看发生动态变化(在实际的k8s中也是这样)，但是存储的路径保证兼容原来的。所以在我们的minik8s中，我们同样借鉴了这样的思路。

最后，特别感谢[github.com/gin-gonic/gin](https://github.com/gin-gonic/gin) 为我们提供了简洁易实现的go语言的Web Server框架。更多详细的接口问题，请查看下面的表格。


#### API接口表

所有的接口都是严格的Restful接口。

| ID                                                      | 请求类型 | URI                                                             | 描述             | 参数说明               | 资源类型       | URI字段                   | 期望返回值       | 字段名 |
| ------------------------------------------------------- | ---- | --------------------------------------------------------------- | -------------- | ------------------ | ---------- | ----------------------- | ----------- | --- |
| [1](https://www.wolai.com/a61XvFRTxGoZ1CbtTiVjQv "1")   | GET  | /api/v1/nodes                                                   | 获取所有的Node      | 暂无                 | Node       | NodesURL                | 200 OK      |     |
| [2](https://www.wolai.com/pwL86UpKfTQLJajW6iUFCt "2")   | POST | /api/v1/nodes                                                   | 创建一个Node       | 暂无                 | Node       | NodesURL                | 201 Created |     |
| [3](https://www.wolai.com/rvA5Yq9H65AGPccXBG5jJk "3")   | GET  | /api/v1/nodes/**:name**                                         | 获取一个Node的信息    | name是节点的名字，必须      | Node       | NodeSpecURL             | 200 OK      |     |
| [5](https://www.wolai.com/hUSNQWagzLi6gV1JbyCWCT "5")   | DEL  | /api/v1/nodes/**:name**                                         | 从集群删除Node      | name是节点的名字，必须      | Node       | NodeSpecURL             | 204 DEL     |     |
| [4](https://www.wolai.com/fMkXyozT24npQnbA3Z4FgE "4")   | PUT  | /api/v1/nodes/**:name**                                         | 更新一个Node的信息    | name是节点的名字，必须      | Node       | NodeSpecURL             | 200 OK      |     |
| [6](https://www.wolai.com/doAbA4DBvrCiUnjj7swozC "6")   | GET  | /api/v1/nodes/**:name**/status                                  | 获取Node的状态      | name是节点的名字，必须      | Node       | NodeSpecStatusURL       | 200 OK      |     |
| [7](https://www.wolai.com/pxUwemQWXVCcVZnmZ7YMJs "7")   | PUT  | /api/v1/nodes/**:name**/status                                  | 更新Node的状态      | name是节点的名字，必须      | Node       | NodeSpecStatusURL       | 200 OK      |     |
| [8](https://www.wolai.com/2YR5VRPCQsDmoeVWsfHyZW "8")   |      |                                                                 |                |                    |            |                         |             |     |
| [9](https://www.wolai.com/cCoWjswopjbMaqDC8YS8Vj "9")   | GET  | /api/v1/namespaces/**:namespace**/pods                          | 获取所有的Pod       | namespace名字空间      | Pod        | PodsURL                 | 200 OK      |     |
| [10](https://www.wolai.com/nWqko2J932btmAcZRkJDpo "10") | POST | /api/v1/namespaces/**:namespace**/pods                          | 创建一个Pod        | namespace名字空间      | Pod        | PodsURL                 | 201 Created |     |
| [11](https://www.wolai.com/iwPJFtsxE6qpTrpsKcgpHs "11") | GET  | /api/v1/namespaces/**:namespace**/pods/**:name**                | 获取某个特定的Pod     | 同上，name是pod名字      | Pod        | PodSpecURL              | 200 OK      |     |
| [12](https://www.wolai.com/4cGomPLTgDcFTCfRvgE6kF "12") | PUT  | /api/v1/namespaces/**:namespace**/pods/**:name**                | 更新某个特定的Pod     | 同上，name是pod名字      | Pod        | PodSpecURL              | 200 OK      |     |
| [13](https://www.wolai.com/m5vxY9cZcaj4Dw831LVtL8 "13") | DEL  | /api/v1/namespaces/**:namespace**/pods/**:name**                | 删除某个Pod        | 同上，name是pod名字      | Pod        | PodSpecURL              | 204 DEL     |     |
| [14](https://www.wolai.com/cFNKxh3pbehfJGwFkfj9XN "14") | GET  | /api/v1/namespaces/**:namespace**/pods/**:name**/status         | 获取某个Pod的状态     | 同上，name是pod名字      | Pod        | PodSpecStatusURL        | 200 OK      |     |
| [15](https://www.wolai.com/gTQSxPJvtB689fAcvUx4qc "15") | POST | /api/v1/namespaces/**:namespace**/pods/**:name**/status         | 更新某个Pod的状态     | 同上，name是pod名字      | Pod        | PodSpecStatusURL        | 200 OK      |     |
| [\_](https://www.wolai.com/gj3QvrPpuQYNRdAV8G1nnv "_")  |      |                                                                 |                |                    |            |                         |             |     |
| [\_](https://www.wolai.com/7w6su8ehDCmsUsFiyGNZ6s "_")  | POST | /api/v1/namespaces/**:namespace**/services                      | 创建一个Service    | namespace名字空间      | Service    | ServiceURL              | 201 Created |     |
| [\_](https://www.wolai.com/om9Uickax7rR4HqXHc4Duc "_")  | GET  | /api/v1/namespaces/**:namespace**/services                      | 获取所有的Service   | namespace名字空间      | Service    | ServiceURL              | 200 OK      |     |
| [\_](https://www.wolai.com/rNzLcTL1wN74umJBnQkJFt "_")  | GET  | /api/v1/namespaces/**:namespace**/services/**:name**            | 获取一个特定Service  | 同上，name是service名字  | Service    | ServiceSpecURL          | 200 OK      |     |
| [\_](https://www.wolai.com/kaMkY6YjTQ7SDDaUWiJYzb "_")  | DEL  | /api/v1/namespaces/**:namespace**/services/**:name**            | 删除一个特定Service  | 同上，name是service名字  | Service    | ServiceSpecURL          | 204 DEL     |     |
| [\_](https://www.wolai.com/tHm2SjYcNCs2Ve2VN1bMxR "_")  | PUT  | /api/v1/namespaces/**:namespace**/services/**:name**            | 更新一个特定Service  | 同上，name是service名字  | Service    | ServiceSpecURL          | 200 OK      |     |
| [\_](https://www.wolai.com/pJHvhdBN3BaeJDSkGvMXUy "_")  |      |                                                                 |                |                    |            |                         |             |     |
| [\_](https://www.wolai.com/kWcXdg48ep7ymRMimKVD73 "_")  | GET  | /apis/v1/namespaces/**:namespace**/jobs                         | 获取所有的Jobs      | namespace名字空间      | Job        | JobsURL                 | 200 OK      |     |
| [\_](https://www.wolai.com/d6LLznhQR12UoA9JujiebX "_")  | POST | /apis/v1/namespaces/**:namespace**/jobs                         | 创建一个Job        | namespace名字空间      | Job        | JobsURL                 | 201 Created |     |
| [\_](https://www.wolai.com/9gxdFDuKceTNBn1TwjcFi2 "_")  | GET  | /apis/v1/namespaces/**:namespace**/jobs/**:name**               | 获取一个特定Job      | 同上，name是Job名字      | Job        | JobSpecURL              | 200 OK      |     |
| [\_](https://www.wolai.com/fJxTqn4Lao8THsrryt9miz "_")  | DEL  | /apis/v1/namespaces/**:namespace**/jobs/**:name**               | 删除一个特定Job      | 同上，name是Job名字      | Job        | JobSpecURL              | 204 DEL     |     |
| [\_](https://www.wolai.com/rH6ExA8PhrhEtmWgPQibcK "_")  | GET  | /apis/v1/namespaces/**:namespace**/jobs/**:name**/status        | 获取一个Job的状态     | 同上，name是Job名字      | Job        | JobSpecStatusURL        | 200 OK      |     |
| [\_](https://www.wolai.com/nZrjxGYY9Vu47TVQrXUwR7 "_")  | PUT  | /apis/v1/namespaces/**:namespace**/jobs/**:name**/status        | 更新一个Job的状态     | 同上，name是Job名字      | Job        | JobSpecStatusURL        | 200 OK      |     |
| [\_](https://www.wolai.com/Ncsdt6P6uURSERYRUGsCn "_")   | POST | /apis/v1/namespaces/**:namespace**/jobfiles                     | 创建一个JobFile    | namespace名字空间      | JobFile    | JobFileURL              | 201 Created |     |
| [\_](https://www.wolai.com/geb7exyffEAM3Jk6QyQrfA "_")  | GET  | /apis/v1/namespaces/**:namespace**/jobfiles                     | 获取一个JobFile内容  | 同上，name是Job名字      | JobFile    | JobFileSpecURL          | 200 OK      |     |
| [\_](https://www.wolai.com/2HBQRXQtwB7qq9NUVCUJZs "_")  | PUT  | /apis/v1/namespaces/**:namespace**/jobfiles/**:name**           | 更新一个JobFile内容  | 同上，name是Job名字      | JobFile    | JobFileSpecURL          | 200 OK      |     |
| [\_](https://www.wolai.com/eTd8g4HpxotCSvXPCMMsfs "_")  |      |                                                                 |                |                    |            |                         |             |     |
| [\_](https://www.wolai.com/j6KPJ1WPFtmBEQLzJ6HT5k "_")  | GET  | /apis/v1/replicasets                                            | 获取全局所有Replica  | 无                  | replicaSet | GlobalReplicaSetsURL    | 200 OK      |     |
| [\_](https://www.wolai.com/pqeGruSbRDtXy1PrBsu3NK "_")  | GET  | /apis/v1/namespaces/**:namespace**/replicasets                  | 获取所有的Replica   | namespace名字空间      | replicaSet | ReplicaSetsURL          | 200 OK      |     |
| [\_](https://www.wolai.com/cY9x3QBKXFX7ZkYbMb25yR "_")  | GET  | /apis/v1/namespaces/**:namespace**/replicasets/**:name**        | 获取特定的Replica   | 同上，name是Job名字      | replicaSet | ReplicaSetSpecURL       | 200 OK      |     |
| [\_](https://www.wolai.com/sUZggecPcPYUq4a8aBrGVW "_")  | DEL  | /apis/v1/namespaces/**:namespace**/replicasets/**:name**        | 删除特定的Replica   | 同上，name是Job名字      | replicaSet | ReplicaSetSpecURL       | 204 DEL     |     |
| [\_](https://www.wolai.com/pDz85LCdb37g2gCdAX9718 "_")  | PUT  | /apis/v1/namespaces/**:namespace**/replicasets/**:name**        | 更新特定的Replica   | 同上，name是Job名字      | replicaSet | ReplicaSetSpecURL       | 200 OK      |     |
| [\_](https://www.wolai.com/nZtS7a1GkZpq5ZvFh5YMmV "_")  | GET  | /apis/v1/namespaces/**:namespace**/replicasets/**:name**/status | 获取特定的Replica   | 同上，name是Job名字      | replicaSet | ReplicaSetSpecStatusURL | 200 OK      |     |
| [\_](https://www.wolai.com/jvNGNJbqMWX9yNj9L8zxp6 "_")  | PUT  | /apis/v1/namespaces/**:namespace**/replicasets/**:name**/status | 更新特定的Replica状态 | 同上，name是Job名字      | replicaSet | ReplicaSetSpecStatusURL | 200 OK      |     |
| [\_](https://www.wolai.com/dqpyD78KwjpkDESaJD1BTk "_")  |      |                                                                 |                |                    |            |                         |             |     |
| [\_](https://www.wolai.com/toGuEZMToEcR2TZvGYyzSa "_")  | GET  | /apis/v1/namespaces/:**namespace**/dns                          | 获取DNS          | namespace名字空间      | DNS        | DnsURL                  | 200 OK      |     |
| [\_](https://www.wolai.com/vPXUfCPp35Gha1bPWALKr4 "_")  | GET  | /apis/v1/namespaces/**:namespace**/dns/**:name**                | 获取DNS          | 同上，name是DNS名字      | DNS        | DnsSpecURL              | 200 OK      |     |
| [\_](https://www.wolai.com/beJvigf6caEe2ndwytsK7p "_")  | POST | /apis/v1/namespaces/:**namespace**/dns                          | 创建DNS          | namespace名字空间      | DNS        | DnsSpecURL              | 201 Created |     |
| [\_](https://www.wolai.com/pUigLuY9EgtMfXVhQsFBYi "_")  | DEL  | /apis/v1/namespaces/**:namespace**/dns/**:name**                | 删除DNS          | 同上，name是DNS名字      | DNS        | DnsSpecURL              | 204 DEL     |     |
| [\_](https://www.wolai.com/6Beh12Kzp17HAy2wDVy43w "_")  |      |                                                                 |                |                    |            |                         |             |     |
| [\_](https://www.wolai.com/nBBgDBniKx4PxgWxY1AkNn "_")  | GET  | /apis/v1/namespaces/:**namespace**/hpa                          | 获取所有的HPA       | namespace名字空间      | HPA        | HPAURL                  | 200 OK      |     |
| [\_](https://www.wolai.com/xoNvxdGW3RyR4ahWWgdMtk "_")  | POST | /apis/v1/namespaces/:**namespace**/hpa                          | 创建HPA          | namespace名字空间      | HPA        | HPAURL                  | 201 Created |     |
| [\_](https://www.wolai.com/3zoSPASfbthu4iizM9nk2L "_")  | GET  | /apis/v1/namespaces/:**namespace**/hpa/**:name**                | 获取特定的HPA       | 同上，name是HPA名字      | HPA        | HPASpecURL              | 200 OK      |     |
| [\_](https://www.wolai.com/gTY1LeWSKbQpF6vW9vM68X "_")  | PUT  | /apis/v1/namespaces/:**namespace**/hpa/**:name**                | 更新特定的HPA       | 同上，name是HPA名字      | HPA        | HPASpecURL              | 200 OK      |     |
| [\_](https://www.wolai.com/dA8jSLmuC9Y334EGhtdhN7 "_")  | DEL  | /apis/v1/namespaces/:**namespace**/hpa/**:name**                | 删除特定的HPA       | 同上，name是HPA名字      | HPA        | HPASpecURL              | 204 DEL     |     |
| [\_](https://www.wolai.com/kV97ZwgBSFYQQG4VNXsP5L "_")  | GET  | /apis/v1/hpa                                                    |                | 无                  | HPA        | GlobalHPAURL            |             |     |
| [\_](https://www.wolai.com/tP9izbcEsrLNTG1fgLwjEP "_")  |      |                                                                 |                |                    |            |                         |             |     |
| [\_](https://www.wolai.com/gR4sBL7HhYsvctc4iBaQ9W "_")  | GET  | /apis/v1/namespaces/:**namespace**/functions                    | 获取所有函数         | namespace名字空间      | Function   | FunctionURL             | 200 OK      |     |
| [\_](https://www.wolai.com/iLqbj5pxZ8TkT9bwA7uVdo "_")  | POST | /apis/v1/namespaces/:**namespace**/functions                    | 创建一个函数         | namespace名字空间      | Function   | FunctionURL             | 201 Created |     |
| [\_](https://www.wolai.com/5yBmMj9gphDnHLjTaSHTnm "_")  | GET  | /apis/v1/namespaces/:**namespace**/functions/:**name**          | 获取特定的函数        | 同上，name是Func名字     | Function   | FunctionSpecURL         | 200 OK      |     |
| [\_](https://www.wolai.com/9Lnk59Xs7FSSU6u8N9sAxk "_")  | PUT  | /apis/v1/namespaces/:**namespace**/functions/:**name**          | 更新特定的函数        | 同上，name是Func名字     | Function   | FunctionSpecURL         | 200 OK      |     |
| [\_](https://www.wolai.com/fwtvNSXrr4cEvumdgP8UDn "_")  | DEL  | /apis/v1/namespaces/:**namespace**/functions/:**name**          | 删除特定的函数        | 同上，name是Func名字     | Function   | FunctionSpecURL         | 204 DEL     |     |
| [\_](https://www.wolai.com/prHZbFfhaEbVdDs99w5HR7 "_")  | GET  | /apis/v1/functions                                              | 获取全局的函数        | 无                  | Function   | GlobalFunctionsURL      | 200 OK      |     |
| [\_](https://www.wolai.com/cUT8xeHi4qcBg5E8oz6btU "_")  |      |                                                                 |                |                    |            |                         |             |     |
| [\_](https://www.wolai.com/cTJfavgFBkBHNbye3dNVGM "_")  | GET  | /apis/v1/namespaces/:**namespace**/workflow                     | 获取某个Workflow   | namespace名字空间      | Workflow   | WorkflowURL             | 200 OK      |     |
| [\_](https://www.wolai.com/hM46SArX29d3r9KZfUY47p "_")  | POST | /apis/v1/namespaces/:**namespace**/workflow                     | 创建一个Workflow   | namespace名字空间      | Workflow   | WorkflowURL             | 201 Created |     |
| [\_](https://www.wolai.com/rW3SdxZYM15cCgbeSYPoPp "_")  | GET  | /apis/v1/namespaces/:**namespace**/workflow/**:name**           | 获取特定Workflow   | 同上，name是workflow名字 | Workflow   | WorkflowSpecURL         | 200 OK      |     |
| [\_](https://www.wolai.com/3j1uzBwkoY7NRPcxdBzPCq "_")  | PUT  | /apis/v1/namespaces/:**namespace**/workflow/**:name**           | 更新Workflow     | 同上，name是workflow名字 | Workflow   | WorkflowSpecURL         | 200 OK      |     |
| [\_](https://www.wolai.com/pr4T5eyzW8HQoV5ifYdwSK "_")  | DEL  | /apis/v1/namespaces/:**namespace**/workflow/**:name**           | 删除Workflow     | 同上，name是workflow名字 | Workflow   | WorkflowSpecURL         | 200 OK      |     |
| [\_](https://www.wolai.com/gWNUphPUYdFzs2TcRje37k "_")  | GET  | /apis/v1/namespaces/:**namespace**/workflow/**:name**/status    | 获取Workflow状态   | 同上，name是workflow名字 | Workflow   | WorkflowSpecStatusURL   | 204 DEL     |     |
| [\_](https://www.wolai.com/eHz22qVpM9gLf4zb7PQcBH "_")  | PUT  | /apis/v1/namespaces/:**namespace**/workflow/**:name**/status    | 更新状态           | 同上，name是workflow名字 | Workflow   | WorkflowSpecStatusURL   | 200 OK      |     |
| [\_](https://www.wolai.com/5C5orAp1widZxekfCeoerc "_")  |      |                                                                 |                |                    |            |                         | 200 OK      |     |

#### 返回的json字段

| 返回字段                                                              | 说明                 |
| ----------------------------------------------------------------- | ------------------ |
| [data](https://www.wolai.com/mRAKDsYrEe6rwVq2dPkg35 "data")       | 返回的数据，可能是数组，也可能是实体 |
| [error](https://www.wolai.com/wVhAKqULJJWRJSWcD5AFyo "error")     | 操作出现错误的原因          |
| [message](https://www.wolai.com/82fbJSmCyzbmE2yfa8uv9k "message") | 操作执行的结果，一般是成功的信息   |



