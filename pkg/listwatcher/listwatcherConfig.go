package listwatcher

import "miniK8s/pkg/message"

type ListwatcherConfig struct {
	subscriberConfig *message.MsgConfig
}

func DefaultListwatcherConfig() *ListwatcherConfig {
	config := ListwatcherConfig{
		subscriberConfig: message.DefaultMsgConfig(),
	}
	return &config
}
