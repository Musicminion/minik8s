package handlers

import (
	"miniK8s/pkg/apiserver/config"
	"miniK8s/pkg/etcd"
	"miniK8s/pkg/k8log"
)

var EtcdStore *etcd.Store = nil

func init() {
	etcdConfig := config.DefaultEtcdConfig()
	etcdStore, err := etcd.NewEtcdStore(etcdConfig.EtcdEndpoints, etcdConfig.EtcdTimeout)
	if err != nil {
		k8log.FatalLog("APIServer", "init etcd client failed, err is "+err.Error())
	}
	k8log.InfoLog("APIServer", "init etcd client connect success")
	EtcdStore = etcdStore
}
