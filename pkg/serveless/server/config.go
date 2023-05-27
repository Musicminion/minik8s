package server

import "time"

var (
	RouterUpdate_Delay    = 0 * time.Second
	RouterUpdate_WaitTime = []time.Duration{10 * time.Second}
	RouterUpdate_ifLoop   = true
)
