package entity

import "miniK8s/pkg/apiObject"

type PodUpdate struct {
	Action string
	PodTarget apiObject.PodStore
	Node   string
	Cmd    []string
}