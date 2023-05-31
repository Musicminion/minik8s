package status

import "time"

const (
	// 缓存在Redis里面的数据库的ID编号，用于区分不同的缓存数据库
	CacheDBID_PodCache = 0
)

// push node status
var (
	NodeHeartBeatInterval = []time.Duration{30 * time.Second}
	NodeHeartBeatDelay    = 0 * time.Second
	NodeHeartBeatLoop     = true
)

// push pod status
var (
	PodStatusUpdateDelay    = 0 * time.Second
	PodStatusUpdateInterval = []time.Duration{5 * time.Second}
	PodStatusUpdateLoop     = true
)

// pull pod item
var (
	PodPullDelay    = 0 * time.Second
	PodPullInterval = []time.Duration{10 * time.Second}
	PodPullIfLoop   = true
)

var (
	PodPushDelay    = 0 * time.Second
	PodPushInterval = []time.Duration{5 * time.Second}
	PodPushIfLoop   = true
)
