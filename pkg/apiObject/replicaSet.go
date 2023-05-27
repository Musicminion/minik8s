package apiObject

import "time"

// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#replicaset-v1-apps

type ReplicaSetSpec struct {
	Replicas int                `json:"replicas" yaml:"replicas"` // 代表副本数
	Selector ReplicaSetSelector `json:"selector" yaml:"selector"` // 代表选择器
	Template PodTemplate        `json:"template" yaml:"template"` // 代表模板
}

type ReplicaSetSelector struct {
	MatchLabels map[string]string `json:"matchLabels" yaml:"matchLabels"` // 代表标签
}

type ReplicaSet struct {
	Basic `json:",inline" yaml:",inline"`
	Spec  ReplicaSetSpec `json:"spec" yaml:"spec"`
}

func (r *ReplicaSet) GetReplicaSetName() string {
	return r.Metadata.Name
}

func (r *ReplicaSet) GetReplicaSetNamespace() string {
	return r.Metadata.Namespace
}

type PodTemplate struct {
	Metadata Metadata `json:"metadata" yaml:"metadata"`
	Spec     PodSpec  `json:"spec" yaml:"spec"`
}

type ReplicaSetStore struct {
	Basic  `json:",inline" yaml:",inline"`
	Spec   ReplicaSetSpec   `json:"spec" yaml:"spec"`
	Status ReplicaSetStatus `json:"status" yaml:"status"`
}

type ReplicaSetStatus struct {
	Replicas      int                   `json:"replicas" yaml:"replicas"`           // 代表期望的副本数
	ReadyReplicas int                   `json:"readyReplicas" yaml:"readyReplicas"` // 代表就绪的副本数
	Conditions    []ReplicaSetCondition `json:"conditions" yaml:"conditions"`       // 代表条件
}

type ReplicaSetCondition struct {
	Type           string    `json:"type" yaml:"type"`                     // 代表条件类型
	Status         string    `json:"status" yaml:"status"`                 // 代表条件状态
	LastUpdateTime time.Time `json:"lastUpdateTime" yaml:"lastUpdateTime"` // 代表最后一次更新时间
	Reason         string    `json:"reason" yaml:"reason"`                 // 代表原因
	Message        string    `json:"message" yaml:"message"`               // 代表消息
}

// 定义ReplicaSet转化为ReplicaSetStore的函数
func (r *ReplicaSet) ToReplicaSetStore() *ReplicaSetStore {
	// 创建一个Status是空的ReplicaSetStore
	return &ReplicaSetStore{
		Basic:  r.Basic,
		Spec:   r.Spec,
		Status: ReplicaSetStatus{},
	}
}

// 定义ReplicaSetStore转化为ReplicaSet的函数
func (r *ReplicaSetStore) ToReplicaSet() *ReplicaSet {
	// 创建一个Status是空的ReplicaSetStore
	return &ReplicaSet{
		Basic: r.Basic,
		Spec:  r.Spec,
	}
}

// 以下函数用来实现apiObject.Object接口
func (r *ReplicaSet) GetObjectKind() string {
	return r.Kind
}

func (r *ReplicaSet) GetObjectName() string {
	return r.Metadata.Name
}

func (r *ReplicaSet) GetObjectNamespace() string {
	return r.Metadata.Namespace
}
