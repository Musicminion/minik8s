package netutil

import (
	"net"
	"sync"
)

var lock sync.Mutex

// 获取一个可以用的端口
func GetAvailablePort() (string, error) {
	lock.Lock()
	defer lock.Unlock()

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}

	defer listener.Close()

	address := listener.Addr().String()

	_, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", err
	}

	return port, nil
}
