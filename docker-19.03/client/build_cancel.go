package client // import "docker-19.03/client"

import (
	"context"
	"net/url"
)

// BuildCancel进程取消正在进行的构建请求
func (cli *Client) BuildCancel(ctx context.Context, id string) error {
	query := url.Values{}
	query.Set("id", id)

	serverResp, err := cli.post(ctx, "/build/cancel", query, nil, nil)
	ensureReaderClosed(serverResp)
	return err
}
