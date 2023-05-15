package entity

import "miniK8s/pkg/apiObject"

type ServiceUpdate struct {
	Action        string
	ServiceTarget apiObject.ServiceStore
}
