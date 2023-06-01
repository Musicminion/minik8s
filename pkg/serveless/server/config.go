package server

import "time"

var (
	RouterUpdate_Delay    = 0 * time.Second
	RouterUpdate_WaitTime = []time.Duration{5 * time.Second}
	RouterUpdate_ifLoop   = true
)
