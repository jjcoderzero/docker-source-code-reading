package httputils // import "docker-19.03/api/server/httputils"

import (
	"io"

	"docker-19.03/api/types/container"
	"docker-19.03/api/types/network"
)

// ContainerDecoder指定如何翻译一个io.Reader进入容器配置。
type ContainerDecoder interface {
	DecodeConfig(src io.Reader) (*container.Config, *container.HostConfig, *network.NetworkingConfig, error)
	DecodeHostConfig(src io.Reader) (*container.HostConfig, error)
}
