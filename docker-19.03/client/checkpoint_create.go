package client // import "docker-19.03/client"

import (
	"context"

	"docker-19.03/api/types"
)

// CheckpointCreate用给定的名称从给定的容器创建一个检查点
func (cli *Client) CheckpointCreate(ctx context.Context, container string, options types.CheckpointCreateOptions) error {
	resp, err := cli.post(ctx, "/containers/"+container+"/checkpoints", nil, options, nil)
	ensureReaderClosed(resp)
	return err
}
