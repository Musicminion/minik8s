package entity

import "miniK8s/pkg/apiObject"

type DnsUpdate struct {
	Action    string
	DnsTarget apiObject.HpaStore
}
