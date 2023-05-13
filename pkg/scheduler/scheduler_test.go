package scheduler

import (
	"reflect"
	"testing"
)

// func TestGetAllNode(t *testing.T) {
// 	code, res, err := netrequest.GetRequest("http://localhost:8090/api/v1/nodes")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	if code != 200 {
// 		t.Error("code is not 200")
// 	}

// 	t.Log(res["data"])

// 	// res["data"]转化为字符串
// 	t.Log()

// 	dataStr := fmt.Sprint(res["data"])
// 	t.Log(dataStr)

// 	var nodes []apiObject.NodeStore

// 	err = json.Unmarshal([]byte(dataStr), &nodes)

// 	if err != nil {
// 		t.Error(err)
// 	}

// }

// func TestTmp(t *testing.T) {
// 	var nodes []apiObject.NodeStore

// 	code, err := netrequest.GetRequestByTarget("http://localhost:8090/api/v1/nodes", &nodes, "data")

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	if code != 200 {
// 		t.Error("code is not 200")
// 	}

// 	// 遍历nodes
// 	for _, node := range nodes {
// 		t.Log(node.GetAPIVersion())
// 	}
// }

func TestNewScheduler(t *testing.T) {
	tests := []struct {
		name    string
		want    *Scheduler
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewScheduler()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewScheduler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewScheduler() = %v, want %v", got, tt.want)
			}
		})
	}
}
