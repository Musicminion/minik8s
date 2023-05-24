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
}
