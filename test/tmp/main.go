package main

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	data := []string{"1", "2"}
	// data := []int{1, 2, 3}

	jsonDatas, err := json.Marshal(data)

	if err != nil {
		fmt.Println(err)
	}

	jsonDatasStr := string(jsonDatas)
	fmt.Println(jsonDatasStr)

	// 使用 gin.H 生成包含 error 和 data 两个键的 JSON 数据
	result := gin.H{
		"data": jsonDatasStr,
	}

	fmt.Println(result)

}
