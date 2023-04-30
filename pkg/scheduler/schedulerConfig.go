package scheduler

type SchedulerConfig struct {
	// 调度策略
	Policy SchedulePolicy
	// apiServer的地址
	ApiServerHost string
	// apiServer的端口
	ApiServerPort int
}

func DefaultSchedulerConfig() *SchedulerConfig {
	config := SchedulerConfig{
		Policy:        RoundRobin,
		ApiServerHost: "localhost",
		ApiServerPort: 8090,
	}
	return &config
}
