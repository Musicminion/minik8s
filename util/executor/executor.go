package executor

import "time"

type callback func()

// waitTime: 等待时间的数组，如果你传入的数组是[1,2,3] 那么程序会等待1s,2s,3s之后再执行callback
// callback: 你需要执行的函数
// ifLoop: 是否循环执行，如果为true，那么程序会一直执行callback, 还是上面的例子，如果ifLoop为true，
// 那么程序会等待1s,2s,3s之后再执行callback，然后再等待1s,2s,3s之后再执行callback，如此无限制的循环下去
// 如果waitTime的长度为0，那么程序会直接返回
// [阻塞函数]：这个函数会阻塞当前的goroutine，如果你想要异步执行，那么请使用go关键字
// callback函数的执行时间不计算在waitTime中
func Period(delay time.Duration, waitTime []time.Duration, callback callback, ifLoop bool) {
	// 为了提高精度，这里使用time.Ticker
	if len(waitTime) == 0 {
		return
	}

	if ifLoop {
		for {
			for _, v := range waitTime {
				callback()
				<-time.After(v)
			}
		}
	} else {
		for _, v := range waitTime {
			callback()
			<-time.After(v)
		}
	}
}

func Delay(waitTime time.Duration, callback callback) {
	<-time.After(waitTime)
	callback()
}
