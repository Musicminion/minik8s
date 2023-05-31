package apiObject

import (
	"time"
)

// 注释By zzq：
// 注意这个API对象有两个，一个是面向用户写的Node，一个是面向etcd存储的NodeStore
// 为什么设计两个呢？因为etcd存储的NodeStore需要存储一些额外的信息，比如运行时的状态
// 而面向用户的Node不需要，这也是体现期望和状态分离的设计思想
// 期望是静态的数据，状态是动态的数据，期望和状态分离是K8s的设计思想
// 为了方便两个对象的转化，我把两个对象的设计的基本都是相似的，除了状态
// 可以调用相关的状态函数，把一个对象转化为另一个对象

// UUID：
// API对象的UUID是不会变的，自从被创建之后，UUID就不会变了，无论编辑还是什么
// 除非重新创建了一个新的对象，这个对象的UUID才会变

// Node
// Node没有Namespace字段，因为Node是集群级别的资源，不属于任何Namespace
// 下面定义的这个Node是面向用户的，而不是存贮到etcd里面的Node
type NodeMetadata struct {
	UUID        string            `json:"uuid" yaml:"uuid"`
	Name        string            `json:"name" yaml:"name"`
	Labels      map[string]string `json:"labels" yaml:"labels"`
	Annotations map[string]string `json:"annotations" yaml:"annotations"`
}

type NodeBasic struct {
	APIVersion   string       `json:"apiVersion" yaml:"apiVersion"`
	Kind         string       `json:"kind" yaml:"kind"`
	NodeMetadata NodeMetadata `json:"metadata" yaml:"metadata"`
}

type Node struct {
	NodeBasic `json:",inline" yaml:",inline"`
	 IP       string `json:"ip" yaml:"ip"`
}

func (n *Node) GetIP() string {
	return n.IP
}

func (n *Node) GetAPIVersion() string {
	return n.NodeBasic.APIVersion
}

func (n *Node) GetUUID() string {
	return n.NodeBasic.NodeMetadata.UUID
}

func (n *Node) GetLabels() map[string]string {
	return n.NodeBasic.NodeMetadata.Labels
}

func (n *Node) GetAnnotations() map[string]string {
	return n.NodeBasic.NodeMetadata.Annotations
}

// 以下函数用来实现apiObject.Object接口
func (n *Node) GetObjectKind() string {
	return n.Kind
}

func (n *Node) GetObjectName() string {
	return n.NodeMetadata.Name
}

func (n *Node) GetObjectNamespace() string {
	// Node没有Namespace
	return ""
}

// 定义Node转化为NodeStore的函数
func (n *Node) ToNodeStore() *NodeStore {
	// 创建一个Status是空的NodeStore
	return &NodeStore{
		NodeBasic: n.NodeBasic,
		IP:        n.IP,
		Status:    NodeStatus{},
	}
}

// 参考K8s官方定义 https://kubernetes.io/zh-cn/docs/concepts/architecture/nodes/#condition
// [边界问题警告！] 如果字符串是空的，那说明这个类型没有设置！
// Ready	如节点是健康的并已经准备好接收 Pod 则为 True；False 表示节点不健康而且不能接收 Pod；
// Unknown 表示节点控制器在最近 node-monitor-grace-period 期间（默认 40 秒）没有收到节点的消息
// DiskPressure	True 表示节点存在磁盘空间压力，即磁盘可用量低, 否则为 False
// MemoryPressure	True 表示节点存在内存压力，即节点内存可用量低，否则为 False
// PIDPressure	True 表示节点存在进程压力，即节点上进程过多；否则为 False
// NetworkUnavailable	True 表示节点网络配置不正确；否则为 False(基本上断联了就是这个情况)

type NodeCondition string

const (
	Ready              NodeCondition = "Ready"
	Unknown            NodeCondition = "Unknown"
	DiskPressure       NodeCondition = "DiskPressure"
	MemoryPressure     NodeCondition = "MemoryPressure"
	PIDPressure        NodeCondition = "PIDPressure"
	NetworkUnavailable NodeCondition = "NetworkUnavailable"
)

// NodeStatus是面向etcd存储的Node!
type NodeStatus struct {
	Hostname   string        `json:"hostname" yaml:"hostname"`
	Ip         string        `json:"ip" yaml:"ip"`
	Condition  NodeCondition `json:"condition" yaml:"condition"`
	CpuPercent float64       `json:"cpuPercent" yaml:"cpuPercent"`
	MemPercent float64       `json:"memPercent" yaml:"memPercent"`
	NumPods    int           `json:"numPods" yaml:"numPods"`
	UpdateTime time.Time     `json:"updateTime" yaml:"updateTime"`
}

// 存储在etcd里面的Node
// 这个NodeStore是面向etcd存储的，而不是面向用户的
type NodeStore struct {
	NodeBasic `json:",inline" yaml:",inline"`
	IP        string     `json:"ip" yaml:"ip"`
	Status    NodeStatus `json:"status" yaml:"status"`
}

// 定义NodeStore到Node的转换函数
func (ns *NodeStore) ToNode() *Node {
	return &Node{
		NodeBasic: ns.NodeBasic,
		IP:        ns.IP,
	}
}

// 定义NodeStore的相关函数,便于获取NodeStore的相关信息
func (ns *NodeStore) GetIP() string {
	return ns.IP
}

func (ns *NodeStore) GetAPIVersion() string {
	return ns.NodeBasic.APIVersion
}

func (ns *NodeStore) GetKind() string {
	return ns.NodeBasic.Kind
}

func (ns *NodeStore) GetUUID() string {
	return ns.NodeBasic.NodeMetadata.UUID
}

func (ns *NodeStore) GetName() string {
	return ns.NodeBasic.NodeMetadata.Name
}

func (ns *NodeStore) GetLabels() map[string]string {
	return ns.NodeBasic.NodeMetadata.Labels
}

func (ns *NodeStore) GetAnnotations() map[string]string {
	return ns.NodeBasic.NodeMetadata.Annotations
}

func (ns *NodeStore) GetStatusHostname() string {
	return ns.Status.Hostname
}

func (ns *NodeStore) GetStatusIp() string {
	return ns.Status.Ip
}

func (ns *NodeStore) GetStatusCondition() NodeCondition {
	return ns.Status.Condition
}

func (ns *NodeStore) GetStatusCpuPercent() float64 {
	return ns.Status.CpuPercent
}

func (ns *NodeStore) GetStatusMemPercent() float64 {
	return ns.Status.MemPercent
}

func (ns *NodeStore) GetStatusNumPods() int {
	return ns.Status.NumPods
}

func (ns *NodeStore) GetStatusUpdateTime() time.Time {
	return ns.Status.UpdateTime
}

// 定义NodeStatus的比较函数，因为要处理Put请求的时候，需要比较两个NodeStatus是否相等

func (ns *NodeStatus) Equal(ns2 *NodeStatus) bool {
	if ns.Hostname != ns2.Hostname {
		return false
	}
	if ns.Ip != ns2.Ip {
		return false
	}
	if ns.Condition != ns2.Condition {
		return false
	}
	if ns.CpuPercent != ns2.CpuPercent {
		return false
	}
	if ns.MemPercent != ns2.MemPercent {
		return false
	}
	if ns.NumPods != ns2.NumPods {
		return false
	}
	if ns.UpdateTime != ns2.UpdateTime {
		return false
	}

	return true
}
