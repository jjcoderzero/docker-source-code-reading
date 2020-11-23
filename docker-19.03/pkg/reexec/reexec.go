package reexec // import "github.com/docker/docker/pkg/reexec"

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var registeredInitializers = make(map[string]func())

// Register在指定的名称下添加初始化函数
func Register(name string, initializer func()) {
	if _, exists := registeredInitializers[name]; exists {
		panic(fmt.Sprintf("reexec func already registered under name %q", name))
	}

	registeredInitializers[name] = initializer
}

// Init作为exec进程的第一部分调用，如果调用了初始化函数，则返回true。
func Init() bool {
	initializer, exists := registeredInitializers[os.Args[0]]
	if exists {
		initializer()

		return true
	}
	return false
}

func naiveSelf() string {
	name := os.Args[0]
	if filepath.Base(name) == name {
		if lp, err := exec.LookPath(name); err == nil {
			return lp
		}
	}
	// 处理相对路径到绝对路径的转换
	if absName, err := filepath.Abs(name); err == nil {
		return absName
	}
	// 如果我们不能得到绝对名称，返回原始(注意:如果操作系统只在Abs()上出错。Getwd失败)
	return name
}
