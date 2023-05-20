package helper

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"miniK8s/pkg/apiserver/app/etcdclient"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	"strconv"
	"strings"
	"time"
)

func AllocClusterIP() (string, error) {
	// 1. 从etcd中获取最大IP
	curMaxIP, err := etcdclient.EtcdStore.Get(serverconfig.EtcdIPPath)
	if err != nil {
		k8log.DebugLog("KUBEPROXY", "allocClusterIP failed, err is "+err.Error())
		return "", err
	}
	// 如果curMaxIP为空，则初始化
	allocatedIP := make(map[int]bool)
	if len(curMaxIP) != 0 {
		// 将对象转为Map
		err = json.Unmarshal([]byte(curMaxIP[0].Value), &allocatedIP)
		if err != nil {
			k8log.DebugLog("KUBEPROXY", "allocClusterIP failed, err is "+err.Error())
			return "", err
		}
	}

	maxTryTime := 100
	num0 := strconv.Itoa(config.SERVICE_IP_PREFIX[0])
	num1 := strconv.Itoa(config.SERVICE_IP_PREFIX[1])

	// 生成两个0-255的随机数
	source := rand.NewSource(time.Now().UnixNano()) // 以当前时间作为随机数种子
	rng := rand.New(source)
	for maxTryTime > 0 {
		maxTryTime--
		// 前三位ip是指定好的
		num2 := (rng.Intn(256))
		num3 := (rng.Intn(256))
		// 判断是否已经分配过
		if allocatedIP[num2*256+num3] {
			continue
		} else {
			allocatedIP[num2*256+num3] = true
			allocatedIPJson, err := json.Marshal(allocatedIP)
			if err != nil {
				k8log.DebugLog("KUBEPROXY", "getAvailableIP failed, err is "+err.Error())
				return "", err
			}
			err = etcdclient.EtcdStore.Put(serverconfig.EtcdIPPath, []byte(allocatedIPJson))
			if err != nil {
				k8log.DebugLog("KUBEPROXY", "getAvailableIP failed, err is "+err.Error())
				return "", err
			}
			return strings.Join([]string{num0, num1, strconv.Itoa(num2), strconv.Itoa(num3)}, "."), nil
		}
	}

	return "", fmt.Errorf("IP is out of range")
}

func JudgeServiceIPAddress(cluserIP string) error {
	// 从etcd中获取IPMap
	curMaxIP, err := etcdclient.EtcdStore.Get(serverconfig.EtcdIPPath)
	if err != nil {
		k8log.DebugLog("KUBEPROXY", "allocClusterIP failed, err is "+err.Error())
		return err
	}
	// 将对象转为Map
	var allocatedIP map[int]bool
	err = json.Unmarshal([]byte(curMaxIP[0].Value), &allocatedIP)
	if err != nil {
		k8log.DebugLog("KUBEPROXY", "allocClusterIP failed, err is "+err.Error())
		return err
	}

	// 将IP拆为IPv4的四个部分
	ipArr := strings.Split(cluserIP, ".")
	if len(ipArr) != 4 {
		k8log.ErrorLog("KUBEPROXY", "allocClusterIP failed, The ipv4 format is invalid")
		return fmt.Errorf("IP is invalid")
	}
	if ipArr[0] != strconv.Itoa(config.SERVICE_IP_PREFIX[0]) || ipArr[1] != strconv.Itoa(config.SERVICE_IP_PREFIX[1]) {
		k8log.ErrorLog("KUBEPROXY", "allocClusterIP failed, The first two addresses of service should be"+
			strconv.Itoa(config.SERVICE_IP_PREFIX[0])+"."+strconv.Itoa(config.SERVICE_IP_PREFIX[1]))
		return fmt.Errorf("IP is invalid")
	}
	num2, _ := strconv.Atoi(ipArr[2])
	num3, _ := strconv.Atoi(ipArr[3])
	if num2 > 255 || num3 > 255 || num2 < 0 || num3 < 0 {
		k8log.ErrorLog("KUBEPROXY", "allocClusterIP failed, The ipv4 format is invalid")
		return fmt.Errorf("IP is invalid")
	}
	if allocatedIP[num2*256+num3] {
		k8log.ErrorLog("KUBEPROXY", "allocClusterIP failed, the address is allocated")
		return fmt.Errorf("IP is allcated")
	}

	// 将IP存入etcd
	allocatedIP[num2*256+num3] = true
	allocatedIPJson, err := json.Marshal(allocatedIP)
	if err != nil {
		k8log.DebugLog("KUBEPROXY", "getAvailableIP failed, err is "+err.Error())
		return err
	}
	err = etcdclient.EtcdStore.Put(serverconfig.EtcdIPPath, []byte(allocatedIPJson))
	if err != nil {
		k8log.DebugLog("KUBEPROXY", "getAvailableIP failed, err is "+err.Error())
		return err
	}
	return nil
}
