package executor

import (
	"fmt"
	"testing"
	"time"
)

func TestPeriod(t *testing.T) {
	waitTimes := []time.Duration{time.Second * 1, time.Second * 2, time.Second * 3, time.Second * 4}
	Period(0, waitTimes, func() {
		fmt.Println("hello world")
		// 打印当前时间
		fmt.Println(time.Now())
	}, false)

}
