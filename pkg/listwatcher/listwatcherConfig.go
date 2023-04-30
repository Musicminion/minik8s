package listwatcher

import "miniK8s/pkg/message"

type listwatcherConfig struct {
	subscriberConfig *message.MsgConfig
}

func DefaultListwatcherConfig() *listwatcherConfig {
	config := listwatcherConfig{
		subscriberConfig: message.DefaultMsgConfig(),
	}
	return &config
}
