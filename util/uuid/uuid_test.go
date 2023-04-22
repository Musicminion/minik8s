// 测试UUID的生成
package uuid

import (
	"fmt"
	"testing"
)

// 在短时间内生成大量的UUID，同时判断是否重复
func TestGetUUID(t *testing.T) {
	// 创建一个containerdClient
	fmt.Printf("TestGetUUID Start")
	// 创建一个字符串到int的map
	uuidMap := make(map[string]int)

	// 循环生成若干次UUID
	for i := 0; i < 10240; i++ {
		uuid := NewUUID()
		// 判断是否为空
		if uuid == "" {
			t.Errorf("GetUUID() failed: %s\n", uuid)
		}else{
			fmt.Printf("uuid generated is: %s\n", uuid)
		}
		// 判断是否重复
		if _, ok := uuidMap[uuid]; ok {
			t.Errorf("GetUUID() duplicated, it is: %s\n", uuid)
		}
		// 把这个字符串放在map中
		uuidMap[uuid] = i
	}
	fmt.Printf("Test Passed\n")
}