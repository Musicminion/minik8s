package function

import "time"

var (
	FuncControllerUpdateDelay     = 0 * time.Second
	FuncControllerUpdateFrequency = []time.Duration{10 * time.Second}
	FuncControllerUpdateLoop      = true
)
