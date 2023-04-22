// 测试令牌生成的处理程序
package handlers

import (
	"testing"
)

// 创建一个全局map，用于存储生成的token
var tokenMap = make(map[string]bool)
var testTime = 16

func TestDelAllToken(t *testing.T) {
	// 1. 删除所有的token
	// handlers.DelAllToken()
	err := DelAllToken()
	// 2. 判断是否删除成功
	if err != nil {
		t.Errorf("delete all token error: %v", err)
	}
}

func TestAddNewToken(t *testing.T) {
	// 用for循环生成1000次token
	for i := 0; i < testTime; i++ {
		// 生成一个token
		token, err := AddNewToken()
		if err != nil {
			t.Errorf("add new token error: %v", err)
		}
		// 判断token是否已经存在
		if _, ok := tokenMap[token]; ok {
			// 如果token已经存在，则报错
			t.Errorf("token %s already exists", token)
		} else {
			// 如果token不存在，则将token存储到map中
			tokenMap[token] = true
			t.Logf(token)
		}
	}
}

func TestVerifyToken(t *testing.T) {
	// 用for循环验证1000次token
	for token := range tokenMap {
		// 验证token
		isExist, err := VerifyToken(token)
		if err != nil {
			t.Errorf("verify token error: %v", err)
		}
		if !isExist {
			t.Errorf("token %s not exists", token)
		}
	}
}

func TestDelToken(t *testing.T) {
	// 用for循环删除1000次token
	for token := range tokenMap {
		// 删除token
		err := DelToken(token)
		if err != nil {
			t.Errorf("delete token error: %v", err)
		}
		// 验证token是否已经删除
		isExist, err := VerifyToken(token)
		if err != nil {
			t.Errorf("verify token error: %v", err)
		}
		if isExist {
			t.Errorf("token %s exists", token)
		}
	}
}

func TestDelAllTokenAgain(t *testing.T) {
	// 1. 删除所有的token
	err := DelAllToken()
	// 2. 判断是否删除成功
	if err != nil {
		t.Errorf("delete all token error: %v", err)
	}
}
