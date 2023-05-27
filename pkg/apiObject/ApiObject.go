package apiObject

import (
	"reflect"
	"strings"
)

const (
	PodKind        = "Pod"
	ServiceKind    = "Service"
	DnsKind        = "Dns"
	NodeKind       = "Node"
	JobKind        = "Job"
	ReplicaSetKind = "Replicaset"
	HpaKind        = "Hpa"
	FunctionKind  = "Function"
)

var AllResourceKind =  strings.ToLower("[" + PodKind + "/" + ServiceKind + "/" + DnsKind + "/" + NodeKind + "/" + JobKind + 
						"/" + ReplicaSetKind + "/" + HpaKind + "/" + FunctionKind + "]")

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
	ReplicaSetKind: reflect.TypeOf(&ReplicaSet{}).Elem(),
	HpaKind: reflect.TypeOf(&HPA{}).Elem(),
	FunctionKind: reflect.TypeOf(&Function{}).Elem(),
}
