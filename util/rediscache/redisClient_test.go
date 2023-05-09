package rediscache

import (
	"testing"
)

var testRedisCache RedisCache

type testStruct struct {
	Name string
	Age  int
}

func TestMain(m *testing.M) {
	testRedisCache = NewRedisCache(0)
	testRedisCache.InitCache()

	m.Run()
}

func TestRedisCache_Put(t *testing.T) {
	// 创建一个结构体
	testStruct := testStruct{
		Name: "test",
		Age:  18,
	}

	testRedisCache.Put("test-1", "test-String")
	testRedisCache.Put("test-2", testStruct)
}

func TestRedisCache_Get(t *testing.T) {
	value, err := testRedisCache.Get("test-1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(value)
	newObj := testStruct{}
	obj, err := testRedisCache.GetObject("test-2", &newObj)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(newObj)
	t.Log(obj)
}

func TestRedisCache_Update(t *testing.T) {
	// 创建一个结构体
	testUpdateStruct := testStruct{
		Name: "test",
		Age:  20,
	}

	err := testRedisCache.Update("test-2", testUpdateStruct)
	if err != nil {
		t.Fatal(err)
	}

	resultStruct := testStruct{}
	returnStructPtr, err := testRedisCache.GetObject("test-2", &resultStruct)

	if err != nil {
		t.Fatal(err)
	}

	// 把returnStructPtr这个指针转换成testStruct类型
	returnStruct := *returnStructPtr.(*testStruct)

	t.Log(returnStruct.Age)
	t.Log(returnStruct.Name)
}
