package handlers

import (
	"encoding/json"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/k8log"
	"miniK8s/util/uuid"
	"time"

	"github.com/gin-gonic/gin"
)

// 获取单个Node信息
// 某个特定的Node状态 对应的NodeSpecURL = "/api/v1/nodes/:name"
func GetNode(c *gin.Context) {
	// 尝试解析请求里面的name
	name := c.Param("name")
	// log
	logStr := "GetNode: name = " + name
	k8log.InfoLog("APIServer", logStr)

	// 如果解析成功，返回对应的Node信息
	if name != "" {
		res, err := EtcdStore.PrefixGet(serverconfig.EtcdNodePath + name)
		if err != nil {
			c.JSON(400, gin.H{
				"error": "get node failed " + err.Error(),
			})
			return
		}
		// 没找到
		if len(res) == 0 {
			c.JSON(404, gin.H{
				"error": "get node err, not find node",
			})
			return
		}

		// 处理res，如果发现有多个Node，返回错误
		if len(res) != 1 {
			c.JSON(500, gin.H{
				"error": "get node err, find more than one node",
			})
			return
		}
		// 遍历res，返回对应的Node信息
		targetNode := res[0].Value
		c.JSON(200, gin.H{
			"data": targetNode,
		})
		return
	} else {
		c.JSON(404, gin.H{
			"error": "name is empty",
		})
		return
	}
}

// 获取所有Node信息
func GetNodes(c *gin.Context) {
	res, err := EtcdStore.PrefixGet(serverconfig.EtcdNodePath)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "get nodes failed " + err.Error(),
		})
		return
	}
	// 遍历res，返回对应的Node信息
	var nodes []string
	for _, node := range res {
		nodes = append(nodes, node.Value)
	}
	c.JSON(200, gin.H{
		"data": nodes,
	})
	// c.JSON(200, nodes)
}

// 删除Node信息
func DeleteNode(c *gin.Context) {
	// 尝试解析请求里面的name
	name := c.Params.ByName("name")
	// 如果解析成功，删除对应的Node信息
	if name != "" {
		// log
		logStr := "DeleteNode: name = " + name
		k8log.InfoLog("APIServer", logStr)

		err := EtcdStore.Del(serverconfig.EtcdNodePath + name)
		if err != nil {
			c.JSON(400, gin.H{
				"error": "delete node failed " + err.Error(),
			})
			return
		}
		c.JSON(204, gin.H{
			"message": "delete node success",
		})
		return
	} else {
		c.JSON(404, gin.H{
			"error": "name is empty",
		})
		return
	}
}

// 添加新的Node
func AddNode(c *gin.Context) {
	// log
	k8log.InfoLog("APIServer", "AddNode: add new node")
	// POST请求，获取请求体
	var node apiObject.Node
	if err := c.ShouldBind(&node); err != nil {
		c.JSON(500, gin.H{
			"error": "parser node failed " + err.Error(),
		})

		k8log.ErrorLog("APIServer", "AddNode: parser node failed "+err.Error())
		return
	}

	// 检查name是否重复
	res, err := EtcdStore.PrefixGet(serverconfig.EtcdNodePath + node.NodeMetadata.Name)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "get node failed " + err.Error(),
		})
		k8log.ErrorLog("APIServer", "AddNode: get node failed "+err.Error())
		return
	}

	if len(res) != 0 {
		c.JSON(500, gin.H{
			"error": "node name already exist",
		})
		k8log.ErrorLog("APIServer", "AddNode: node name already exist")
		return
	}
	// 检查Node的kind是否正确
	if node.Kind != "Node" {
		c.JSON(500, gin.H{
			"error": "node kind is not Node",
		})
		k8log.ErrorLog("APIServer", "AddNode: node kind is not Node")
		return
	}

	// 给Node设置UUID, 所以哪怕用户故意设置UUID也会被覆盖
	node.NodeMetadata.UUID = uuid.NewUUID()

	// 将Node转化为NodeStore
	nodeStore := node.ToNodeStore()

	// 把nodeStore转化为json
	nodeJson, err := json.Marshal(nodeStore)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "node marshal to json failed" + err.Error(),
		})
		return
	}

	// 将Node信息写入etcd
	err = EtcdStore.Put(serverconfig.EtcdNodePath+node.NodeMetadata.Name, nodeJson)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "put node to etcd failed" + err.Error(),
		})
		return
	}
	// 返回201处理成功
	c.JSON(201, gin.H{
		"message": "create node success",
	})
}

// 更新Node信息
// 对于一些数组类型的变量，我们采用覆盖的方式，而不是append的方式
// 比如原来是key1-val1、key2-val2，现在POST的是key3-val3，那么最后的结果就是key3-val3，原来的两个键值对被删除
// 所以用户如果要追加，需要自行通过GET获取信息然后再PUT！这种扔给调用者自己处理
func UpdateNode(c *gin.Context) {
	// 这个是PUT请求，解析请求URI里面的name
	name := c.Params.ByName("name")
	if name != "" {
		// log
		logStr := "UpdateNode: name = " + name
		k8log.InfoLog("APIServer", logStr)

		// 先获取原来的Node信息
		res, err := EtcdStore.PrefixGet(serverconfig.EtcdNodePath + name)
		if err != nil {
			k8log.DebugLog("APIServer", "UpdateNode: get node failed "+err.Error())
			c.JSON(400, gin.H{
				"error": "get node failed " + err.Error(),
			})
			return
		}

		// 处理res，如果发现有多个Node，返回错误
		if len(res) != 1 {
			k8log.DebugLog("APIServer", "UpdateNode: find more than one node")
			c.JSON(500, gin.H{
				"error": "get node err, find more than one node",
			})
			return
		}

		// 把POST请求里面的Node信息解析出来
		newNode := apiObject.NodeStore{}
		if err := c.ShouldBind(&newNode); err != nil {
			k8log.DebugLog("APIServer", "UpdateNode: parser post node failed "+err.Error())
			c.JSON(500, gin.H{
				"error": "parser post node failed " + err.Error(),
			})
			return
		}

		// 把原来的Node信息解析出来
		oldNode := apiObject.NodeStore{}
		err = json.Unmarshal([]byte(res[0].Value), &oldNode)
		if err != nil {
			k8log.DebugLog("APIServer", "UpdateNode: unmarshal old node failed "+err.Error())
			c.JSON(500, gin.H{
				"error": "unmarshal old node failed " + err.Error(),
			})
			return
		}

		// 选择性的更新Node信息
		selectiveUpdateNode(&oldNode, &newNode)

		// 把更新后的Node信息转化为json
		nodeJson, err := json.Marshal(oldNode)
		if err != nil {
			k8log.DebugLog("APIServer", "UpdateNode: marshal node failed "+err.Error())
			c.JSON(500, gin.H{
				"error": "marshal node failed " + err.Error(),
			})
			return
		}

		// 把更新后的Node信息写入etcd
		err = EtcdStore.Put(serverconfig.EtcdNodePath+name, nodeJson)
		if err != nil {
			k8log.DebugLog("APIServer", "UpdateNode: put node to etcd failed "+err.Error())
			c.JSON(500, gin.H{
				"error": "put node to etcd failed " + err.Error(),
			})
			return
		}

		// 返回200处理成功
		c.JSON(200, gin.H{
			"message": "update node success",
			"data":    oldNode,
		})

	} else {
		c.JSON(404, gin.H{
			"error": "name is empty",
		})
		return
	}

}

// 选择性更新Node的字段，不是所有的字段都可以更新
func selectiveUpdateNode(oldNode *apiObject.NodeStore, postNode *apiObject.NodeStore) {
	// Node不是想更新什么就更新什么的，有些字段是不允许更新的
	// 只有Status字段是允许更新，还有Labels Annotations

	// 遍历执行在oldNode上面执行更新
	// Labels处理
	if len(postNode.NodeMetadata.Labels) != 0 {
		// 先清空原来的Labels
		oldNode.NodeMetadata.Labels = make(map[string]string)
		// 然后根据POST的Node信息更新
		for key, value := range postNode.NodeMetadata.Labels {
			oldNode.NodeMetadata.Labels[key] = value
		}
	}

	// Annotations处理
	if len(postNode.NodeMetadata.Annotations) != 0 {
		// 先清空原来的Annotations
		oldNode.NodeMetadata.Annotations = make(map[string]string)
		// 然后根据POST的Node信息更新
		for key, value := range postNode.NodeMetadata.Annotations {
			oldNode.NodeMetadata.Annotations[key] = value
		}
	}

	// Status处理，
	// 创建一个空白的Status，然后根据比较
	emptyStatue := apiObject.NodeStatus{}
	// 如果新的Status不为初始化的默认值，那么就更新
	if postNode.Status != emptyStatue {
		if postNode.Status.Hostname != "" {
			oldNode.Status.Hostname = postNode.Status.Hostname
		}
		if postNode.Status.Ip != "" {
			oldNode.Status.Ip = postNode.Status.Ip
		}
		if postNode.Status.Condition != "" {
			oldNode.Status.Condition = postNode.Status.Condition
		}
		if postNode.Status.CpuPercent != 0 {
			oldNode.Status.CpuPercent = postNode.Status.CpuPercent
		}
		if postNode.Status.MemPercent != 0 {
			oldNode.Status.MemPercent = postNode.Status.MemPercent
		}
		if postNode.Status.NumPods != 0 {
			oldNode.Status.NumPods = postNode.Status.NumPods
		}
		// 根据当前时间更新UpdateTime
		oldNode.Status.UpdateTime = time.Now()
	}
}
