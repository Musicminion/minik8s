package containerdClient
import (
	"testing"
	"fmt"
)

// func (c *containerdClient) TestgetClient (t *testing.T) {
// 	fmt.Println("client")
// 	client, err := c.getClient()
// 	if err != nil {
// 		t.Errorf("getClient() failed: %v", err)
// 	}
// 	t.Logf("client: %v", client)
// }

func TestGetClient(t *testing.T) {
	// 创建一个containerdClient
	fmt.Println("TestGetClientCreate")
	c := &containerdClient{}
	client, err := c.GetClient()
	if err != nil {
		t.Errorf("GetClient() failed: %v", err)
	}
	if client == nil {
		t.Errorf("GetClient() failed: %v", err)
	}
	fmt.Println("Test Passed")
}

func TestCloseClient(t *testing.T) {
	// 创建一个containerdClient
	fmt.Println("TestCloseClient")
	c := &containerdClient{}
	c.GetClient()
	err := c.CloseClient()
	if err != nil {
		t.Errorf("CloseClient() failed: %v", err)
	}
	fmt.Println("Test Passed")
}

