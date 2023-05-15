package main

import (
	apiserver "miniK8s/pkg/apiserver/app/server"
	"miniK8s/pkg/apiserver/serverconfig"
)

func main() {
	// apiServer := apiserver.New(config.DefaultServerConfig())
	// apiServer.Run()

	apiserver.New(serverconfig.DefaultServerConfig()).Run()
}
