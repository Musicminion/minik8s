package scheduler

import (
	"encoding/json"
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/apiserver/app/etcdclient"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/k8log"
	"miniK8s/util/uuid"
	"testing"
)

var testNode1 = apiObject.Node{
	NodeBasic: apiObject.NodeBasic{
		APIVersion: "v1",
		Kind:       "Node",
		NodeMetadata: apiObject.NodeMetadata{
			Name: "node1",
			UUID: uuid.NewUUID(),
		},
	},
	IP: "192.168.1.1",
}

var testNode2 = apiObject.Node{
	NodeBasic: apiObject.NodeBasic{
		APIVersion: "v1",
		Kind:       "Node",
		NodeMetadata: apiObject.NodeMetadata{
			Name: "node2",
			UUID: uuid.NewUUID(),
		},
	},
	IP: "192.168.2.1",
}

func TestGetAllNodes(t *testing.T) {

	// 将testNode1 marshal
	testNodeBytes, err := json.Marshal(testNode1)
	if err != nil {
		t.Error(err)
	}

	// 向etcd中添加node1
	err = etcdclient.EtcdStore.Put(serverconfig.EtcdNodePath+testNode1.GetName(), (testNodeBytes))
	if err != nil {
		t.Error(err)
	}

	// 将testNode2 marshal
	testNodeBytes, err = json.Marshal(testNode2)
	if err != nil {
		t.Error(err)
	}

	// 向etcd中添加node2
	err = etcdclient.EtcdStore.Put(serverconfig.EtcdNodePath+testNode2.GetName(), (testNodeBytes))
	if err != nil {
		t.Error(err)
	}

	sch := &Scheduler{
		apiServerHost: "127.0.0.1",
		apiServerPort: 8090,
	}

	nodes, err := sch.GetAllNodes()

	k8log.InfoLog("scheduler", "nodes: "+fmt.Sprint(nodes))

	if err != nil {
		t.Error(err)
	}

	// 验证node的信息是否正确

	if nodes == nil || len(nodes) == 0 {
		t.Error("nodes is nil or len is wrong")
	}

}
