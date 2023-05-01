package minik8stypes

import "github.com/docker/go-connections/nat"

type PortBinding struct {
	HostIp   string // Host IP address that the container's port is mapped to
	HostPort string // Host port number that the container's port is mapped to
}

// ContainerType表示容器的类型
// 参考Docker官方的API https://github.com/docker/go-docker/blob/master/api/types/container/config.go
// 编写一个容器的配置结构体，在创建容器的时候使用
type ContainerConfig struct {
	// 所有的Container相关的配置都在这里
	// ************************************************ //
	Tty bool     // 是否需要Tty终端 Attach standard streams to a tty, including stdin if it is not closed.
	Env []string // 环境变量 List of environment variable to set in the container
	// Cmd和Entrypoint的区别：
	// https://stackoverflow.com/questions/21553353/what-is-the-difference-between-cmd-and-entrypoint-in-a-dockerfile
	// Entrypoint是容器启动的时候执行的命令，而Cmd是在启动容器的时候传递给Entrypoint的参数
	// 例如，Entrypoint是bash，Cmd是-c，那么最终执行的命令就是bash -c
	// Cmd是可以被docker run的命令行参数覆盖的，而Entrypoint是不可以的
	Cmd             []string            // 启动子容器的时候执行的命令 Command to run when starting the container
	Entrypoint      []string            // Entrypoint to run when starting the container
	Image           string              // Name of the image as it was passed by the operator (e.g. could be symbolic)
	ImagePullPolicy ImagePullPolicy     `default:"IfNotPresent"` // 拉取镜像的策略, Always, Never, IfNotPresent
	Volumes         map[string]struct{} // List of volumes (mounts) used for the container
	Labels          map[string]string   // List of labels set to this container
	ExposedPorts    map[string]struct{} // List of exposed ports

	// ************************************************ //
	// 所有的HostConfig相关的配置都在这里
	VolumesFrom  []string    // List of volumes to take from other containers
	Links        []string    // List of links (in the name:alias form)
	NetworkMode  string      // [网络模式] Network mode to use for the container
	PidMode      string      // [PidMode] PID namespace to use for the container
	IpcMode      string      // [IPC Mode ]IPC namespace to use for the container(设置这三个可以让容器共享主机的网络、PID、IPC、NetworkMode)
	Binds        []string    // List of volume bindings for this container
	PortBindings nat.PortMap // List of port bindings for this container
	// ************************************************ //

	// 下面的是Docker官方的API中的字段，用作开发参考
	// Hostname        string              // Hostname
	// Domainname      string              // Domainname
	// User            string              // User that will run the command(s) inside the container, also support user:group
	// AttachStdin     bool                // Attach the standard input, makes possible user interaction
	// AttachStdout    bool                // Attach the standard output
	// AttachStderr    bool                // Attach the standard error
	// ExposedPorts    nat.PortSet         // List of exposed ports
	// Tty             bool                // Attach standard streams to a tty, including stdin if it is not closed.
	// OpenStdin       bool                // Open stdin
	// StdinOnce       bool                // If true, close stdin after the 1 attached client disconnects.
	// Env             []string            // List of environment variable to set in the container
	// Cmd             strslice.StrSlice   // Command to run when starting the container
	// Healthcheck     *HealthConfig       // Healthcheck describes how to check the container is healthy
	// ArgsEscaped     bool                // True if command is already escaped (Windows specific)
	// Image           string              // Name of the image as it was passed by the operator (e.g. could be symbolic)
	// Volumes         map[string]struct{} // List of volumes (mounts) used for the container
	// WorkingDir      string              // Current directory (PWD) in the command will be launched
	// Entrypoint      strslice.StrSlice   // Entrypoint to run when starting the container
	// NetworkDisabled bool                // Is network disabled
	// MacAddress      string              // Mac Address of the container
	// OnBuild         []string            // ONBUILD metadata that were defined on the image Dockerfile
	// Labels          map[string]string   // List of labels set to this container
	// StopSignal      string              // Signal to stop a container
	// StopTimeout     *int                // Timeout (in seconds) to stop a container
	// Shell           strslice.StrSlice   // Shell for shell-form of RUN, CMD, ENTRYPOINT
}

// 根据官方API的定义，这里我用ContainerStatus表示容器的状态
// 原始的注解如下：Status string
// String representation of the container state.
// Can be one of "created", "running", "paused", "restarting", "removing", "exited", or "dead"
type ContainerStatus string

const (
	Created  ContainerStatus = "created"
	Running  ContainerStatus = "running"
	Paused   ContainerStatus = "paused"
	Restart  ContainerStatus = "restarting"
	Removing ContainerStatus = "removing"
	Exited   ContainerStatus = "exited"
	Dead     ContainerStatus = "dead"
)

const (
	PORT_PROTOCOL_TCP = "tcp"
	PORT_PROTOCOL_UDP = "udp"
	PORT_LOCALHOST_IP = "127.0.0.1"
)

const (
	Contianer_IPCMode_Sharable = "shareable"
)

// 系统保留字段，给容器的标签使用
const (
	ContainerLabel_Pod = "_pod"
)
