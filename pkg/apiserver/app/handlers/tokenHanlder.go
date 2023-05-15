package handlers

import (
	"errors"
	"math/rand"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/apiserver/app/etcdclient"
	"miniK8s/pkg/k8log"
	"time"
)

// 生成一个长度为64的随机字母和数字的组合的token
func GenerateToken() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	source := rand.NewSource(time.Now().UnixNano())
	var seededRand *rand.Rand = rand.New(source)
	b := make([]byte, 64)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// 返回一个新的令牌，并在屏幕上显示它
func AddNewToken() (string, error) {
	// 1. 生成一个长度为64的随机字母和数字的组合的token
	newToken := GenerateToken()
	// 记录日志
	logStr := "new token created: " + newToken
	k8log.InfoLog("APIServer", logStr)

	// 2. 将token存储到etcd中
	key := serverconfig.EtcdTokenPath + newToken
	etcdclient.EtcdStore.Put(key, []byte(newToken))
	// 3. 返回token
	return newToken, nil
}

func DelToken(token string) error {
	// 1. 删除etcd中的token
	key := serverconfig.EtcdTokenPath + token
	err := etcdclient.EtcdStore.Del(key)
	// 记录日志
	logStr := "token deleted: " + token
	k8log.WarnLog("APIServer", logStr)
	return err
}

func DelAllToken() error {
	// 1. 删除etcd中的所有token
	logStr := "all token deleted!"
	k8log.WarnLog("APIServer", logStr)
	err := etcdclient.EtcdStore.PrefixDel(serverconfig.EtcdTokenPath)
	return err
}

func VerifyToken(token string) (bool, error) {
	// 1. 从etcd中获取token
	key := serverconfig.EtcdTokenPath + token
	res, err := etcdclient.EtcdStore.Get(key)
	if err != nil {
		return false, err
	}
	// 2. 如果token存在，则返回true，否则返回false
	if len(res) == 1 {
		logStr := "token verified successful: " + token
		k8log.InfoLog("APIServer", logStr)
		return true, nil
	} else if len(res) > 1 {
		logStr := "token verified failed[duplicate token error]: " + token
		k8log.InfoLog("APIServer", logStr)
		return false, errors.New("token duplicate error")
	}
	logStr := "token verified failed: " + token
	k8log.InfoLog("APIServer", logStr)
	return false, nil
}
