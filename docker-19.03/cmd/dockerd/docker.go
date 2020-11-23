package main

import (
	"fmt"
	"os"

	"docker-19.03/buildkit/util/apicaps"
	"docker-19.03/cli"
	"docker-19.03/daemon/config"
	"docker-19.03/dockerversion"
	"docker-19.03/pkg/jsonmessage"
	"docker-19.03/pkg/reexec"
	"docker-19.03/pkg/term"
	"docker-19.03/rootless"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	honorXDG bool
)

func newDaemonCommand() (*cobra.Command, error) {
	opts := newDaemonOptions(config.New())

	cmd := &cobra.Command{
		Use:           "dockerd [OPTIONS]",
		Short:         "A self-sufficient runtime for containers.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.flags = cmd.Flags()
			return runDaemon(opts) // 真正的入口
		},
		DisableFlagsInUseLine: true,
		Version:               fmt.Sprintf("%s, build %s", dockerversion.Version, dockerversion.GitCommit),
	}
	cli.SetupRootCommand(cmd)

	flags := cmd.Flags()
	flags.BoolP("version", "v", false, "Print version information and quit")
	defaultDaemonConfigFile, err := getDefaultDaemonConfigFile()
	if err != nil {
		return nil, err
	}
	flags.StringVar(&opts.configFile, "config-file", defaultDaemonConfigFile, "Daemon configuration file")
	opts.InstallFlags(flags)
	if err := installConfigFlags(opts.daemonConfig, flags); err != nil {
		return nil, err
	}
	installServiceFlags(flags)

	return cmd, nil
}

func init() {
	if dockerversion.ProductName != "" {
		apicaps.ExportedProduct = dockerversion.ProductName
	}
	// 在使用RootlessKit运行时，需要将$XDG_RUNTIME_DIR、$XDG_DATA_HOME和$XDG_CONFIG_HOME作为默认dir，因为我们不太可能有权限访问系统范围内的目录。
	// 注意，即使使用——rootless运行，当不使用RootlessKit运行时，honorXDG也需要保持为false，因为当前挂载名称空间中的系统范围目录是可访问的。(“rootful”dockerd在rootless dockerd， #38702)
	honorXDG = rootless.RunningWithRootlessKit()
}

func main() {
	// 在docker运行之前没有进行任何Initializer注册，故代码段执行的返回值为假。
	// reexec存在的作用: 协调execdriver与容器创建时dockerinit这两者的关系。
	if reexec.Init() {
		return
	}

	// initial的日志格式;这个设置在加载守护进程配置之后更新。
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: jsonmessage.RFC3339NanoFixed,
		FullTimestamp:   true,
	})

	// 根据需要设置基于平台的终端仿真。
	_, stdout, stderr := term.StdStreams()

	initLogging(stdout, stderr)

	onError := func(err error) {
		fmt.Fprintf(stderr, "%s\n", err)
		os.Exit(1)
	}

	cmd, err := newDaemonCommand()
	if err != nil {
		onError(err)
	}
	cmd.SetOutput(stdout)
	if err := cmd.Execute(); err != nil {
		onError(err)
	}
}
