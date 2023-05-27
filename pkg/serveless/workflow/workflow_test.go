package workflow

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	// JSON字符串
	jsonStr := `{"a": 42, "b": "hello", "c": [1, 2, 3]}`

	// 解析JSON字符串
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		fmt.Println("解析JSON失败:", err)
		return
	}

	// 打印解析后的数据
	fmt.Println("解析JSON成功:", data)

}
