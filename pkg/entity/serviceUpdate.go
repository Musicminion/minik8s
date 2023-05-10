package entity

import (
)

type ServiceUpdate struct {
	Action string
	ServiceTarget ServiceWithEndpoints
}