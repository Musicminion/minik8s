package apiObject

import "reflect"

const (
	PodKind     = "Pod"
	ServiceKind = "Service"
	DnsKind     = "Dns"
	NodeKind    = "Node"
	JobKind     = "Job"
)

type APIObject interface {
	// GetObjectName() string
	GetObjectKind() string
	GetObjectName() string
	GetObjectNamespace() string
}

// kind -> apiObject
var KindToStructType = map[string]reflect.Type{
	PodKind:     reflect.TypeOf(&Pod{}).Elem(),
	ServiceKind: reflect.TypeOf(&Service{}).Elem(),
	DnsKind:     reflect.TypeOf(&Dns{}).Elem(),
	JobKind:     reflect.TypeOf(&Job{}).Elem(),
	NodeKind:    reflect.TypeOf(&Node{}).Elem(),
}
