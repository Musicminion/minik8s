package runtime

import (
	"miniK8s/pkg/apiObject"
	"testing"
)

var testPod apiObject.PodStore

func TestAll(t *testing.T) {
	runtimeManager := NewRuntimeManager()
	runtimeManager.CreatePod(&testPod)

	runtimeManager.DeletePod(&testPod)

}
