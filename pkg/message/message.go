package message

import "encoding/json"

// 规定Message通信的格式
const (
	// 请求调度,消息内容为Pod的API地址，如/api/v1/namespaces/default/pods/mypod
	// 处理者（调度器）会主动去拉取Pod的信息，然后发送信息给API Server的消息队列
	// APIServer会watch自己的消息队列，然后接收到请求后处理
	RequestSchedule = "RequestSchedule"

	// 调度结果,消息内容为Pod被调度到的节点的名称
	ScheduleResult = "ScheduleResult"
)

const (
	CREATE string = "CREATE"
	DELETE string = "DELETE"
	UPDATE string = "UPDATE"
	EXEC   string = "EXEC"
)

type Message struct {
	// 消息对应事件的类型,用来区别不同的事件
	Type string `json:"type"`
	// 消息内容
	Content string `json:"content"`
	// 资源URI
	ResourceURI string `json:"resourceURI"`
	// 资源名称
	ResourceName string `json:"resourceName"`
}

// 将Message转换为Json格式的Message（从string转换）
func ParseJsonMessageFromStr(msg string) (*Message, error) {
	var result Message
	err := json.Unmarshal([]byte(msg), &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// 将Message转换为Json格式的Message（从[]byte转换）
func ParseJsonMessageFromBytes(msg []byte) (*Message, error) {
	var result Message
	err := json.Unmarshal(msg, &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}
