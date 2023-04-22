package handlers

import (
	"miniK8s/pkg/apiserver/config"
	"miniK8s/pkg/etcd"
)

var EtcdStore *etcd.Store = nil

func init() {
	etcdConfig := config.DefaultEtcdConfig()
	etcdStore, err := etcd.NewEtcdStore(etcdConfig.EtcdEndpoints, etcdConfig.EtcdTimeout)
	if err != nil {
		panic(err)
	}
	EtcdStore = etcdStore
}
