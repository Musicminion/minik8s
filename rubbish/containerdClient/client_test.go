// 测试containerdClient
package containerdClient

import (
	"fmt"
	"testing"
)

// 测试containerdClient
func TestNewContainerdClient(t *testing.T) {
	client, err := NewContainerdClient()
	if err != nil {
		fmt.Printf("TestContainerdClient failed: %v\n", err)
	}
	defer client.Close()
	fmt.Printf("TestContainerdClient passed\n")
}
