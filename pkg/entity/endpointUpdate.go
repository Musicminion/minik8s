package entity

import (
)

type EndpointUpdate struct {
	Action string
	ServiceTarget ServiceWithEndpoints
}