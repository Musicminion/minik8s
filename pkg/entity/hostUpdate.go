package entity

import "miniK8s/pkg/apiObject"

type HostUpdate struct {
	Action    string
	DnsTarget apiObject.HpaStore
	DnsConfig string
	HostList  []string
}
