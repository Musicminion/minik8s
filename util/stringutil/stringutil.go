package stringutil

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// GenerateRandomStr 生成给定长度的随机字符串
func GenerateRandomStr(length int) string {
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rng.Intn(len(letterRunes))]
	}
	return string(b)
}

// replace 将 URL 中的 toChangeStr 替换为newStr
func Replace(URL string, toChangeStr string, newStr string) string {
	return strings.Replace(URL, toChangeStr, newStr, -1)
}

func StringSliceToJsonArray(data []string) string {
	jsonDatas := strings.Join(data, ",")
	return fmt.Sprint("[", jsonDatas, "]")
}

func ContainsString(list []string, target string) bool {
    for _, str := range list {
    	if str == target {
            return true
        }
    }
    return false
}
