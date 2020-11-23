package client // import "docker-19.03/client"

import (
	"context"
	"encoding/json"
	"net/url"

	"docker-19.03/api/types"
)

// CheckpointList返回docker主机中给定容器的检查点
func (cli *Client) CheckpointList(ctx context.Context, container string, options types.CheckpointListOptions) ([]types.Checkpoint, error) {
	var checkpoints []types.Checkpoint

	query := url.Values{}
	if options.CheckpointDir != "" {
		query.Set("dir", options.CheckpointDir)
	}

	resp, err := cli.get(ctx, "/containers/"+container+"/checkpoints", query, nil)
	defer ensureReaderClosed(resp)
	if err != nil {
		return checkpoints, wrapResponseError(err, resp, "container", container)
	}

	err = json.NewDecoder(resp.body).Decode(&checkpoints)
	return checkpoints, err
}
