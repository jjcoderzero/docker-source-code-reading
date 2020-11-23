package rootless // import "github.com/docker/docker/rootless"

import (
	"os"
	"sync"
)

const (
	// RootlessKitDockerProxyBinary是rootlesskit-docker-proxy的二进制名称
	RootlessKitDockerProxyBinary = "rootlesskit-docker-proxy"
)

var (
	runningWithRootlessKit     bool
	runningWithRootlessKitOnce sync.Once
)

// 如果在RootlessKit名称空间下运行，RunningWithRootlessKit则返回true。
func RunningWithRootlessKit() bool {
	runningWithRootlessKitOnce.Do(func() {
		u := os.Getenv("ROOTLESSKIT_STATE_DIR")
		runningWithRootlessKit = u != ""
	})
	return runningWithRootlessKit
}
