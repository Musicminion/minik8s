package entity

import (
	"miniK8s/pkg/apiObject"
)

type ServiceWithEndpoints struct {
	Endpoints  []apiObject.Endpoint
	Service    apiObject.ServiceStore
}