package container // import "docker-19.03/api/types/container"

import (
	"strings"

	"docker-19.03/api/types/blkiodev"
	"docker-19.03/api/types/mount"
	"docker-19.03/api/types/strslice"
	"docker-19.03/go-connections/nat"
	"docker-19.03/go-units"
)

// Isolation 表示容器的隔离技术。支持的值是特定于平台的
type Isolation string

// IsDefault表示容器的默认隔离技术。在Linux上，这是本机驱动程序。在Windows上，这是一个Windows服务器容器。
func (i Isolation) IsDefault() bool {
	return strings.ToLower(string(i)) == "default" || string(i) == ""
}

// IsHyperV表示使用Hyper-V分区进行隔离
func (i Isolation) IsHyperV() bool {
	return strings.ToLower(string(i)) == "hyperv"
}

// IsProcess表示使用进程隔离
func (i Isolation) IsProcess() bool {
	return strings.ToLower(string(i)) == "process"
}

const (
	// IsolationEmpty是未指定的(与default行为相同)
	IsolationEmpty = Isolation("")
	// IsolationDefault是当前守护进程的默认隔离模式
	IsolationDefault = Isolation("default")
	// IsolationProcess是进程隔离模式
	IsolationProcess = Isolation("process")
	// IsolationHyperV是HyperV隔离模式
	IsolationHyperV = Isolation("hyperv")
)

// IpcMode表示容器的ipc堆栈。
type IpcMode string

// IsPrivate表示容器是否使用它自己的私有ipc名称空间，它不能被共享。
func (n IpcMode) IsPrivate() bool {
	return n == "private"
}

// IsHost表示容器是否共享主机的ipc名称空间。
func (n IpcMode) IsHost() bool {
	return n == "host"
}
//  IsShareable表示容器的ipc名称空间是否可以与另一个容器共享。
func (n IpcMode) IsShareable() bool {
	return n == "shareable"
}

// IsContainer指示容器是否使用另一个容器的ipc名称空间。
func (n IpcMode) IsContainer() bool {
	parts := strings.SplitN(string(n), ":", 2)
	return len(parts) > 1 && parts[0] == "container"
}

// IsNone表示容器的IpcMode是否设置为“none”。
func (n IpcMode) IsNone() bool {
	return n == "none"
}

// IsEmpty表示容器IpcMode是否为空
func (n IpcMode) IsEmpty() bool {
	return n == ""
}

// Valid表示ipc模式是否有效
func (n IpcMode) Valid() bool {
	return n.IsEmpty() || n.IsNone() || n.IsPrivate() || n.IsHost() || n.IsShareable() || n.IsContainer()
}

// Container返回将要使用的容器ipc堆栈的名称。
func (n IpcMode) Container() string {
	parts := strings.SplitN(string(n), ":", 2)
	if len(parts) > 1 && parts[0] == "container" {
		return parts[1]
	}
	return ""
}

// NetworkMode表示容器网络堆栈。
type NetworkMode string

// IsNone表示容器是否没有使用网络堆栈。
func (n NetworkMode) IsNone() bool {
	return n == "none"
}

// IsDefault表示容器是否使用默认网络堆栈。
func (n NetworkMode) IsDefault() bool {
	return n == "default"
}

// IsPrivate表示容器是否使用它的私有网络栈。
func (n NetworkMode) IsPrivate() bool {
	return !(n.IsHost() || n.IsContainer())
}

// IsContainer表示容器是否使用容器网络栈。
func (n NetworkMode) IsContainer() bool {
	parts := strings.SplitN(string(n), ":", 2)
	return len(parts) > 1 && parts[0] == "container"
}

// ConnectedContainer是该容器所连接的网络的容器的id。
func (n NetworkMode) ConnectedContainer() string {
	parts := strings.SplitN(string(n), ":", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// UserDefined表示用户创建的网络
func (n NetworkMode) UserDefined() string {
	if n.IsUserDefined() {
		return string(n)
	}
	return ""
}

// UsernsMode represents userns mode in the container.
type UsernsMode string

// IsHost indicates whether the container uses the host's userns.
func (n UsernsMode) IsHost() bool {
	return n == "host"
}

// IsPrivate indicates whether the container uses the a private userns.
func (n UsernsMode) IsPrivate() bool {
	return !(n.IsHost())
}

// Valid indicates whether the userns is valid.
func (n UsernsMode) Valid() bool {
	parts := strings.Split(string(n), ":")
	switch mode := parts[0]; mode {
	case "", "host":
	default:
		return false
	}
	return true
}

// CgroupSpec表示容器要使用的cgroup。
type CgroupSpec string

// IsContainer表示容器是否正在使用另一个容器cgroup
func (c CgroupSpec) IsContainer() bool {
	parts := strings.SplitN(string(c), ":", 2)
	return len(parts) > 1 && parts[0] == "container"
}

// Valid指示cgroup规范是否有效。
func (c CgroupSpec) Valid() bool {
	return c.IsContainer() || c == ""
}

// Container返回将使用其cgroup的容器的名称。
func (c CgroupSpec) Container() string {
	parts := strings.SplitN(string(c), ":", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// UTSMode表示容器的UTS名称空间。
type UTSMode string

// IsPrivate指示容器是否使用其私有的UTS名称空间。
func (n UTSMode) IsPrivate() bool {
	return !(n.IsHost())
}

// IsHost指示容器是否使用主机的UTS名称空间。
func (n UTSMode) IsHost() bool {
	return n == "host"
}

// Valid表示UTS名称空间是否有效。
func (n UTSMode) Valid() bool {
	parts := strings.Split(string(n), ":")
	switch mode := parts[0]; mode {
	case "", "host":
	default:
		return false
	}
	return true
}

// PidMode表示容器的pid名称空间。
type PidMode string

// IsPrivate表示容器是否使用它自己的新的pid名称空间。
func (n PidMode) IsPrivate() bool {
	return !(n.IsHost() || n.IsContainer())
}

// IsHost指示容器是否使用主机的pid名称空间。
func (n PidMode) IsHost() bool {
	return n == "host"
}

// IsContainer指示容器是否使用容器的pid命名空间。
func (n PidMode) IsContainer() bool {
	parts := strings.SplitN(string(n), ":", 2)
	return len(parts) > 1 && parts[0] == "container"
}

// Valid指示pid名称空间是否有效。
func (n PidMode) Valid() bool {
	parts := strings.Split(string(n), ":")
	switch mode := parts[0]; mode {
	case "", "host":
	case "container":
		if len(parts) != 2 || parts[1] == "" {
			return false
		}
	default:
		return false
	}
	return true
}

// Container返回将要使用其pid命名空间的容器的名称。
func (n PidMode) Container() string {
	parts := strings.SplitN(string(n), ":", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// DeviceRequest表示来自设备驱动程序的设备请求。由GPU设备驱动程序使用。
type DeviceRequest struct {
	Driver       string            // 设备驱动程序名称
	Count        int               // 请求的设备数量(-1 =ALL)
	DeviceIDs    []string          // 由设备驱动程序识别的设备id列表
	Capabilities [][]string        // 设备功能的一个或多个列表(例如:“gpu”)
	Options      map[string]string // 要传递到设备驱动程序的选项
}

// DeviceMapping表示主机和容器之间的设备映射。
type DeviceMapping struct {
	PathOnHost        string
	PathInContainer   string
	CgroupPermissions string
}

// RestartPolicy表示容器的重启策略
type RestartPolicy struct {
	Name              string
	MaximumRetryCount int
}

// IsNone表示容器是否具有“no”重启策略。这意味着退出时容器不会自动重启。
func (rp *RestartPolicy) IsNone() bool {
	return rp.Name == "no" || rp.Name == ""
}

// IsAlways表示容器是否具有“always”重启策略。这意味着无论退出状态如何，容器都将自动重新启动。
func (rp *RestartPolicy) IsAlways() bool {
	return rp.Name == "always"
}

// IsOnFailure指示容器是否具有“on-failure”重启策略。这意味着容器将以非零退出状态自动重新启动退出。
func (rp *RestartPolicy) IsOnFailure() bool {
	return rp.Name == "on-failure"
}

// IsUnlessStopped表示容器是否具有“unless-stopped”重启策略。这意味着容器将自动重新启动，除非用户已将其置于停止状态。
func (rp *RestartPolicy) IsUnlessStopped() bool {
	return rp.Name == "unless-stopped"
}

// IsSame比较两个RestartPolicy，看它们是否相同
func (rp *RestartPolicy) IsSame(tp *RestartPolicy) bool {
	return rp.Name == tp.Name && rp.MaximumRetryCount == tp.MaximumRetryCount
}

// LogMode是一种类型，用于定义用于记录日志的可用模式。当日志消息开始堆积时，这些模式会影响对日志的处理方式。
type LogMode string

// 可用的日志模式
const (
	LogModeUnset            = ""
	LogModeBlocking LogMode = "blocking"
	LogModeNonBlock LogMode = "non-blocking"
)

// LogConfig表示容器的日志配置。
type LogConfig struct {
	Type   string
	Config map[string]string
}

// Resources包含容器的资源(cgroups config, ulimits…)
type Resources struct {
	// Applicable to all platforms
	CPUShares int64 `json:"CpuShares"` // CPU shares (relative weight vs. other containers)
	Memory    int64 // Memory limit (in bytes)
	NanoCPUs  int64 `json:"NanoCpus"` // CPU quota in units of 10<sup>-9</sup> CPUs.

	// Applicable to UNIX platforms
	CgroupParent         string // Parent cgroup.
	BlkioWeight          uint16 // Block IO weight (relative weight vs. other containers)
	BlkioWeightDevice    []*blkiodev.WeightDevice
	BlkioDeviceReadBps   []*blkiodev.ThrottleDevice
	BlkioDeviceWriteBps  []*blkiodev.ThrottleDevice
	BlkioDeviceReadIOps  []*blkiodev.ThrottleDevice
	BlkioDeviceWriteIOps []*blkiodev.ThrottleDevice
	CPUPeriod            int64           `json:"CpuPeriod"`          // CPU CFS (Completely Fair Scheduler) period
	CPUQuota             int64           `json:"CpuQuota"`           // CPU CFS (Completely Fair Scheduler) quota
	CPURealtimePeriod    int64           `json:"CpuRealtimePeriod"`  // CPU real-time period
	CPURealtimeRuntime   int64           `json:"CpuRealtimeRuntime"` // CPU real-time runtime
	CpusetCpus           string          // CpusetCpus 0-2, 0,1
	CpusetMems           string          // CpusetMems 0-2, 0,1
	Devices              []DeviceMapping // List of devices to map inside the container
	DeviceCgroupRules    []string        // List of rule to be added to the device cgroup
	DeviceRequests       []DeviceRequest // List of device requests for device drivers
	KernelMemory         int64           // Kernel memory limit (in bytes)
	KernelMemoryTCP      int64           // Hard limit for kernel TCP buffer memory (in bytes)
	MemoryReservation    int64           // Memory soft limit (in bytes)
	MemorySwap           int64           // Total memory usage (memory + swap); set `-1` to enable unlimited swap
	MemorySwappiness     *int64          // Tuning container memory swappiness behaviour
	OomKillDisable       *bool           // Whether to disable OOM Killer or not
	PidsLimit            *int64          // Setting PIDs limit for a container; Set `0` or `-1` for unlimited, or `null` to not change.
	Ulimits              []*units.Ulimit // List of ulimits to be set in the container

	// Applicable to Windows
	CPUCount           int64  `json:"CpuCount"`   // CPU count
	CPUPercent         int64  `json:"CpuPercent"` // CPU percent
	IOMaximumIOps      uint64 // Maximum IOps for the container system drive
	IOMaximumBandwidth uint64 // Maximum IO in bytes per second for the container system drive
}

// UpdateConfig holds the mutable attributes of a Container.
// Those attributes can be updated at runtime.
type UpdateConfig struct {
	// Contains container's resources (cgroups, ulimits)
	Resources
	RestartPolicy RestartPolicy
}

// 容器的不可移植配置结构。这里，“non-portable”意味着“依赖于我们运行的主机”。可移植信息*应该*出现在配置中。
type HostConfig struct {
	// 适用于所有平台
	Binds           []string      // 此容器的卷绑定列表
	ContainerIDFile string        // 写入containerId的文件(路径)
	LogConfig       LogConfig     // 此容器的日志配置
	NetworkMode     NetworkMode   // 为容器使用的网络模式
	PortBindings    nat.PortMap   // 公开端口(容器)和主机之间的端口映射
	RestartPolicy   RestartPolicy // 用于容器的重新启动策略
	AutoRemove      bool          // 当容器退出时自动删除它
	VolumeDriver    string        // 用于挂载卷的卷驱动程序的名称
	VolumesFrom     []string      // 从其他容器中的卷组列表

	// 适用于UNIX平台
	CapAdd          strslice.StrSlice // 要添加到容器中的内核功能列表
	CapDrop         strslice.StrSlice // 要从容器中删除的内核功能列表
	Capabilities    []string          `json:"Capabilities"` // 可供容器使用的内核功能列表(这覆盖了默认设置)
	DNS             []string          `json:"Dns"`          // 要查找的DNS服务器列表
	DNSOptions      []string          `json:"DnsOptions"`   // 要查找的DNSOption列表
	DNSSearch       []string          `json:"DnsSearch"`    // 要查找的DNSSearch列表
	ExtraHosts      []string          // 额外主机列表
	GroupAdd        []string          // 容器进程将作为其他组运行的列表
	IpcMode         IpcMode           // 用于容器的IPC名称空间
	Cgroup          CgroupSpec        // 用于容器的Cgroup
	Links           []string          // 链接列表(以name:alias形式)
	OomScoreAdj     int               // 容器偏爱OOM-killing
	PidMode         PidMode           // 用于容器的PID名称空间
	Privileged      bool              // 容器是否处于特权模式
	PublishAllPorts bool              // docker应该为容器发布所有暴露的端口吗
	ReadonlyRootfs  bool              // 容器根文件系统是只读的吗
	SecurityOpt     []string          // 用于为MLS系统(如SELinux)定制标签的字符串值列表。
	StorageOpt      map[string]string `json:",omitempty"` // 每个容器的存储驱动程序选项。
	Tmpfs           map[string]string `json:",omitempty"` // 用于容器的tmpfs(挂载)列表
	UTSMode         UTSMode           // 用于容器的UTS名称空间
	UsernsMode      UsernsMode        // 要用于容器的用户名称空间
	ShmSize         int64             // shm内存使用总量
	Sysctls         map[string]string `json:",omitempty"` // 用于容器的具有命名空间的系统名列表
	Runtime         string            `json:",omitempty"` // 与此容器一起使用的运行时

	// Applicable to Windows
	ConsoleSize [2]uint   // 初始控制台大小(高度，宽度)
	Isolation   Isolation // 容器的隔离技术(例如默认的hyperv)

	// 包含容器的资源(cgroups, ulimit)
	Resources

	// 挂载容器使用的规格
	Mounts []mount.Mount `json:",omitempty"`

	// MaskedPaths是要在容器内隐藏的路径列表(这覆盖了默认的路径集)
	MaskedPaths []string

	// ReadonlyPaths是在容器内设置为只读的路径列表(这将覆盖默认的路径集)
	ReadonlyPaths []string

	// 在容器内运行自定义init，如果为空，则使用守护进程配置的设置
	Init *bool `json:",omitempty"`
}
