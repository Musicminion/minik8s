package k8log

import "testing"

var outTimes = 2

func TestInfoLog(t *testing.T) {
	// 循环输出5次
	for i := 0; i < outTimes; i++ {
		InfoLog("test log info: aaaaa")
	}
}

func TestErrorLog(t *testing.T) {
	for i := 0; i < outTimes; i++ {
		ErrorLog("test log error: aaaaa")
	}
}

func TestWarnLog(t *testing.T) {
	for i := 0; i < outTimes; i++ {
		WarnLog("test log warn: aaaaa")
	}
}

func TestDebugLog(t *testing.T) {
	for i := 0; i < outTimes; i++ {
		DebugLog("test log debug: aaaaa")
	}
}

// func TestFatalLog(t *testing.T) {
// 	FatalLog("test log fatal: aaaaa")
// }
