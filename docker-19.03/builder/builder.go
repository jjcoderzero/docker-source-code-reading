// 为任何Docker builder定义要实现的接口。
package builder // import "docker-19.03/builder"

import (
	"context"
	"io"

	"docker-19.03/api/types"
	"docker-19.03/api/types/backend"
	"docker-19.03/api/types/container"
	containerpkg "docker-19.03/container"
	"docker-19.03/image"
	"docker-19.03/layer"
	"docker-19.03/pkg/containerfs"
)

const (
	DefaultDockerfileName = "Dockerfile" // DefaultDockerfileName是Docker命令的默认文件名，由Docker build读取
)

// Source定义了一个位置，它可以用作构建器中ADD/COPY指令的源。
type Source interface {
	Root() containerfs.ContainerFS // Root返回访问源的根路径
	Close() error // Close允许发出文件系统树不再被使用的信号。对于使用临时目录的上下文实现，建议删除Close()中的临时目录。
	Hash(path string) (string, error) // Hash返回文件的校验和
}

// Backend抽象对Docker守护进程的调用。
type Backend interface {
	ImageBackend
	ExecBackend
	CommitBuildStep(backend.CommitConfig) (image.ID, error) // CommitBuildStep从构建步骤生成的配置创建一个新的Docker镜像
	ContainerCreateWorkdir(containerID string) error // ContainerCreateWorkdir创建工作区

	CreateImage(config []byte, parent string) (Image, error)

	ImageCacheBuilder
}

// ImageBackend是镜像组件所需的接口方法
type ImageBackend interface {
	GetImageAndReleasableLayer(ctx context.Context, refOrID string, opts backend.GetImageAndLayerOptions) (Image, ROLayer, error)
}

// ExecBackend包含执行容器所需的接口方法
type ExecBackend interface {
	ContainerAttachRaw(cID string, stdin io.ReadCloser, stdout, stderr io.Writer, stream bool, attached chan struct{}) error // ContainerAttachRaw连接到容器。
	ContainerCreateIgnoreImagesArgsEscaped(config types.ContainerCreateConfig) (container.ContainerCreateCreatedBody, error) // ContainerCreateIgnoreImagesArgsEscaped创建一个新的Docker容器并返回潜在的警告
	ContainerRm(name string, config *types.ContainerRmConfig) error // ContainerRm删除由“id”指定的容器。
	ContainerKill(containerID string, sig uint64) error // ContainerKill突然停止容器执行。
	ContainerStart(containerID string, hostConfig *container.HostConfig, checkpoint string, checkpointDir string) error // ContainerStart启动一个新的容器
	ContainerWait(ctx context.Context, name string, condition containerpkg.WaitCondition) (<-chan containerpkg.StateStatus, error) // ContainerWait停止处理，直到给定容器停止。
}

// Result是构建器生成的输出
type Result struct {
	ImageID   string
	FromImage Image
}

// ImageCacheBuilder表示有状态镜像缓存的生成器。
type ImageCacheBuilder interface {
	// MakeImageCache创建一个有状态镜像缓存。
	MakeImageCache(cacheFrom []string) ImageCache
}

// ImageCache对镜像缓存进行抽象。(parent image, child runconfig) -> child image
type ImageCache interface {
	// GetCache返回一个对缓存映像的引用，该映像的父节点为' parent '， runconfig为' cfg '。缓存未命中会返回一个空ID和一个nil错误。
	GetCache(parentID string, cfg *container.Config) (imageID string, err error)
}

// Image表示构建器使用的Docker镜像。
type Image interface {
	ImageID() string
	RunConfig() *container.Config
	MarshalJSON() ([]byte, error)
	OperatingSystem() string
}

// ROLayer是对镜像rootfs层的引用
type ROLayer interface {
	Release() error
	NewRWLayer() (RWLayer, error)
	DiffID() layer.DiffID
}

// RWLayer是活动层，可以read/modified
type RWLayer interface {
	Release() error
	Root() containerfs.ContainerFS
	Commit() (ROLayer, error)
}
