package message

// K8s消息交换机名字
const K8sExchange = "K8sExchange"

// 可能使用到的ContentType
const ContentTypeJson = "application/json"
const ContentTypeText = "text/plain"

// 消息组件的配置
type MsgConfig struct {
	// RabbitMQ服务器地址
	User     string
	Password string
	Host     string
	Port     int
	// 虚拟Host
	VHost string
	// 最大重连次数 int
	MaxReconnect int
	// 重连间隔时间 s秒
	ReconnectInterval int
}

// 默认的配置是连接本地的RabbitMQ服务器，使用Guest账号
func DefaultMsgConfig() *MsgConfig {
	config := MsgConfig{
		User:     "guest",
		Password: "guest",
		Host:     "localhost",
		Port:     5672,
		VHost:    "/",
	}
	return &config
}
