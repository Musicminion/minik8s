package workflow

import "time"

var (
	WorkflowController_Delay    = 0 * time.Second
	WorkflowController_Waittime = []time.Duration{20 * time.Second}
	WorkflowController_ifLoop   = true
)
