package pleg

import "time"

var (
	PlegFirstRunDelay = 5 * time.Second
	PlegRunPeriod     = []time.Duration{time.Second * 10}
	PlegRunRoutine    = true
)
