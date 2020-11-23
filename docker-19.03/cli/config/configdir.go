package config // import "docker-19.03/cli/config"

import (
	"os"
	"path/filepath"

	"docker-19.03/pkg/homedir"
)

var (
	configDir     = os.Getenv("DOCKER_CONFIG")
	configFileDir = ".docker"
)

// Dir返回DOCKER_CONFIG环境变量指定的配置目录的路径。如果DOCKER_CONFIG未设置，Dir将返回~/docker。Dir忽略XDG_CONFIG_HOME(与docker客户机相同)。
// TODO: this was copied from cli/config/configfile and should be removed once cmd/dockerd moves
func Dir() string {
	return configDir
}

func init() {
	if configDir == "" {
		configDir = filepath.Join(homedir.Get(), configFileDir)
	}
}
