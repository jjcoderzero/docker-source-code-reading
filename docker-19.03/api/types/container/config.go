package container // import "docker-19.0.3/api/types/container"

import (
	"docker-19.03/api/types/strslice"
	"time"

	"docker-19.03/go-connections/nat"
)

// MinimumDuration在用户配置的持续时间上放置一个最小值。这是为了防止时间单元上的API错误。例如，API可能将3设置为健康检查间隔，意图为3秒，但Docker将其解释为3纳秒。
const MinimumDuration = 1 * time.Millisecond

// HealthConfig保存HEALTHCHECK特性的配置设置。
type HealthConfig struct {
	// 测试是用于检查容器是否健康的测试。空的slice意味着继承默认值。
	// The options are:
	// {} : 继承healthcheck
	// {"NONE"} : 禁止healthcheck
	// {"CMD", args...} : 直接执行参数
	// {"CMD-SHELL", command} : 使用系统的默认shell运行命令
	Test []string `json:",omitempty"`

	// Zero意味着继承。持续时间用整数纳秒表示。
	Interval    time.Duration `json:",omitempty"` // Interval是检查之间等待的时间.
	Timeout     time.Duration `json:",omitempty"` // Timeout是在考虑检查是否已挂起之前等待的时间.
	StartPeriod time.Duration `json:",omitempty"` // 重试开始倒数之前容器初始化的开始周期n.

	// Retries是认为容器不健康所需的连续失败次数。零意味着继承。
	Retries int `json:",omitempty"`
}

// Config包含关于容器的配置数据.它应该只保存关于容器的可移植信息
// 这里，“portable”的意思是“独立于我们运行的主机”。HostConfig中应该出现不可移植的信息。添加到这个结构中的所有字段都必须标记为'omitempty'，以从旧的'v1Compatibility'配置中获取可预测的散列。
type Config struct {
	Hostname        string              // 主机名
	Domainname      string              // 域名
	User            string              // 将在容器内运行命令的User，也支持 user:group
	AttachStdin     bool                // 附加标准输入，使用户交互成为可能
	AttachStdout    bool                // 附加标准输出
	AttachStderr    bool                // 附加标准错误输出
	ExposedPorts    nat.PortSet         `json:",omitempty"` // 暴露端口清单
	Tty             bool                // 将标准流附加到tty，包括未关闭的stdin。
	OpenStdin       bool                // 打开stdin
	StdinOnce       bool                // 如果为真，在连接的1个客户端断开连接后关闭stdin。
	Env             []string            // 要在容器中设置的环境变量的列表
	Cmd             strslice.StrSlice   // 命令在启动容器时运行
	Healthcheck     *HealthConfig       `json:",omitempty"` // Healthcheck 描述如何检查容器是否健康
	ArgsEscaped     bool                `json:",omitempty"` // 如果命令已经转义(意味着作为命令行处理)(特定于Windows)，则为真。
	Image           string              // 操作符传递镜像时的名称(例如，可以是符号)
	Volumes         map[string]struct{} // 用于容器的卷(挂载)的列表
	WorkingDir      string              // 当前目录(PWD)中的命令将被启动
	Entrypoint      strslice.StrSlice   // 启动容器时运行的入口点
	NetworkDisabled bool                `json:",omitempty"` // 是网络禁用
	MacAddress      string              `json:",omitempty"` // 容器的Mac地址
	OnBuild         []string            // 在image Dockerfile上定义的ONBUILD元数据
	Labels          map[string]string   // 设置到此容器的标签列表
	StopSignal      string              `json:",omitempty"` // 停止容器的信号
	StopTimeout     *int                `json:",omitempty"` // 超时(以秒为单位)去停止容器
	Shell           strslice.StrSlice   `json:",omitempty"` // Shell for shell-form of RUN, CMD, ENTRYPOINT
}
