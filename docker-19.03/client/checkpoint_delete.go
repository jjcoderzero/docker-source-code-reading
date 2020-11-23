package client // import "docker-19.03/client"

import (
	"context"
	"net/url"

	"docker-19.03/api/types"
)

// CheckpointDelete从给定容器中删除具有给定名称的检查点
func (cli *Client) CheckpointDelete(ctx context.Context, containerID string, options types.CheckpointDeleteOptions) error {
	query := url.Values{}
	if options.CheckpointDir != "" {
		query.Set("dir", options.CheckpointDir)
	}

	resp, err := cli.delete(ctx, "/containers/"+containerID+"/checkpoints/"+options.CheckpointID, query, nil)
	ensureReaderClosed(resp)
	return err
}
