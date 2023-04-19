package test

import (
	// "fmt"
	"github.com/stretchr/testify/assert"
	"miniK8s/pkg/etcd"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	store, err := etcd.NewEtcdStore([]string{"localhost:2379"},5 * time.Second)
	assert.Nil(t, err)
	err = store.Put("my_key1", []byte("my_value1"))
	assert.Nil(t, err)
	val, _ := store.Get("my_key1")
	assert.Equal(t, "my_value1", string(val[0].ValueBytes))
	err = store.Del("my_key1")
	assert.Nil(t, err)
}
