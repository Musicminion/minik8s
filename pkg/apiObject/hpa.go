package apiObject

import "time"

type HPAMetrics struct {
	CPUPercent float64 `yaml:"cpuPercent" json:"cpuPercent"`
	MemPercent float64 `yaml:"memPercent" json:"memPercent"`
}

type HPASpec struct {
	MinReplicas    int           `yaml:"minReplicas" json:"minReplicas"`
	MaxReplicas    int           `yaml:"maxReplicas" json:"maxReplicas"`
	Workload       Basic         `yaml:"workload" json:"workload"`
	AdjustInterval time.Duration `yaml:"adjustInterval" json:"adjustInterval"` // 调整的时间间隔
	Selector       HPASelector   `yaml:"selector" json:"selector"`
	Metrics        HPAMetrics  `yaml:"metrics" json:"metrics"`
}

type HPASelector struct {
	MatchLabels map[string]string `yaml:"matchLabels" json:"matchLabels"`
}

type HPA struct {
	Basic `yaml:",inline" json:",inline"`
	Spec  HPASpec `yaml:"spec" json:"spec"`
}

type HPAStore struct {
	Basic  `yaml:",inline" json:",inline"`
	Spec   HPASpec   `yaml:"spec" json:"spec"`
	Status HPAStatus `yaml:"status" json:"status"`
}

type HPAStatus struct {
	CurrentReplicas int `yaml:"currentReplicas" json:"currentReplicas"`
	CurCPUPercent    float64 `yaml:"curCPUPercent" json:"curCPUPercent"`
	CurMemPercent    float64 `yaml:"curMemPercent" json:"curMemPercent"`
}

// 定义hpa到hpaStore的转换函数
func (hpa *HPA) ToHPAStore() *HPAStore {
	return &HPAStore{
		Basic: hpa.Basic,
		Spec:  hpa.Spec,
	}
}

// 定义hpaStore到hpa的转换函数
func (hs *HPAStore) ToHPA() *HPA {
	return &HPA{
		Basic: hs.Basic,
		Spec:  hs.Spec,
	}
}


// 以下函数用来实现apiObject.Object接口
func (h *HPA) GetObjectKind() string {
	return h.Kind
}

func (h *HPA) GetObjectName() string {
	return h.Metadata.Name
}

func (h *HPA) GetObjectNamespace() string {
	return h.Metadata.Namespace
}