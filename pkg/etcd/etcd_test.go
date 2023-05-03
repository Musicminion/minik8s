// 测试etcd的功能
package etcd

import (
	"testing"
	"time"
)

// 创建全局变量客户端
var testEtcdStore *Store = nil

// 在测试之前，NewEtcdStore
func TestMain(m *testing.M) {
	Store, err := NewEtcdStore([]string{"localhost:2379"}, 5*time.Second)
	if err != nil {
		panic(err)
	}
	testEtcdStore = Store
	testEtcdStore.DelAll()
	m.Run()
}

// 测试Put方法
func TestPut(t *testing.T) {
	err := testEtcdStore.Put("/test/child1", []byte("testchildValue1"))
	if err != nil {
		t.Fatal(err)
	}
	err = testEtcdStore.Put("/test/child2", []byte("testchildValue2"))
	if err != nil {
		t.Fatal(err)
	}
	err = testEtcdStore.Put("/test/child3", []byte("testchildValue3"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestGet(t *testing.T) {
	res, err := testEtcdStore.Get("/test")
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 0 {
		t.Fatal("get error")
	}
	t.Log(res)

	res, err = testEtcdStore.Get("/testNoExist")
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 0 {
		t.Fatal("get error")
	}
	t.Log(res)
}

func TestPrefixGet(t *testing.T) {
	res, err := testEtcdStore.PrefixGet("/test")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}

func TestDel(t *testing.T) {
	err := testEtcdStore.Del("/test/child1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestPrefixDel(t *testing.T) {
	err := testEtcdStore.PrefixDel("/test")
	if err != nil {
		t.Fatal(err)
	}
}
