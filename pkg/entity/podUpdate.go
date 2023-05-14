package entity

import "miniK8s/pkg/apiObject"

type PodUpdate struct {
	Action string
	PodTarget apiObject.Pod
	Node   string
}