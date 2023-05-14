package scheduler

import (
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/listwatcher"
	"miniK8s/pkg/message"
	"reflect"
	"testing"
)

func TestScheduler_GetAllNodes(t *testing.T) {
	type fields struct {
		lw            *listwatcher.Listwatcher
		publisher     *message.Publisher
		polocy        SchedulePolicy
		apiServerHost string
		apiServerPort int
	}
	tests := []struct {
		name      string
		fields    fields
		wantNodes []apiObject.NodeStore
		wantErr   bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			fields: fields{
				lw:            nil,
				publisher:     nil,
				polocy:        "RoundRobin",
				apiServerHost: "localhost",
				apiServerPort: 8090,
			},
			wantNodes: []apiObject.NodeStore{
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sch := &Scheduler{
				lw:            tt.fields.lw,
				publisher:     tt.fields.publisher,
				polocy:        tt.fields.polocy,
				apiServerHost: tt.fields.apiServerHost,
				apiServerPort: tt.fields.apiServerPort,
			}
			gotNodes, err := sch.GetAllNodes()
			k8log.DebugLog("scheduler", "gotNodes: "+gotNodes[len(gotNodes)-1].GetName())

			k8log.DebugLog("scheduler", "gotNodes: ")
			if (err != nil) != tt.wantErr {
				t.Errorf("Scheduler.GetAllNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotNodes, tt.wantNodes) {
				t.Errorf("Scheduler.GetAllNodes() = %v, want %v", gotNodes, tt.wantNodes)
			}
		})
	}
}
