package client

import "net/http"

// NewClient为给定主机和API版本初始化一个新的API客户机。它使用给定的http客户机作为传输。它还初始化要添加到每个请求中的自定义http头。
// 如果版本号为空，则不会发送任何版本信息。强烈建议您设置一个版本，否则如果服务器升级，您的客户端可能会崩溃。弃用:使用NewClientWithOpts
func NewClient(host string, version string, client *http.Client, httpHeaders map[string]string) (*Client, error) {
	return NewClientWithOpts(WithHost(host), WithVersion(version), WithHTTPClient(client), WithHTTPHeaders(httpHeaders))
}

// NewEnvClient根据环境变量初始化一个新的API客户机。有关支持环境变量的列表，请参阅FromEnv。弃用:使用NewClientWithOpts (FromEnv)
func NewEnvClient() (*Client, error) {
	return NewClientWithOpts(FromEnv)
}
