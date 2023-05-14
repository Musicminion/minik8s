package status

import "time"

const (
	// 缓存在Redis里面的数据库的ID编号，用于区分不同的缓存数据库
	CacheDBID_PodCache = 0
)

var (
	NodeHeartBeatInterval = []time.Duration{10 * time.Second}
	NodeHeartBeatDelay    = 0 * time.Second
	NodeHeartBeatLoop     = true
)
