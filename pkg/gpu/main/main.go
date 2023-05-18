package main

import (
	"miniK8s/pkg/gpu/jobserver"
	"time"
)

func main() {
	// jobserver, err := jobserver.NewJobServer(jobserver.NewJobServerConfig())
	// if err != nil {
	// 	panic(err)
	// }
	// if jobserver == nil {
	// 	panic("jobserver is nil")
	// }
	jobserver.NewJobServer(jobserver.NewJobServerConfig())
	time.Sleep(1000 * time.Second)

}
