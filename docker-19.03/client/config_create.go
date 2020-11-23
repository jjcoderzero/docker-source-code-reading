package client // import "docker-19.03/client"

import (
	"context"
	"encoding/json"

	"docker-19.03/api/types"
	"docker-19.03/api/types/swarm"
)

// ConfigCreate创建一个新配置。
func (cli *Client) ConfigCreate(ctx context.Context, config swarm.ConfigSpec) (types.ConfigCreateResponse, error) {
	var response types.ConfigCreateResponse
	if err := cli.NewVersionError("1.30", "config create"); err != nil {
		return response, err
	}
	resp, err := cli.post(ctx, "/configs/create", nil, config, nil)
	defer ensureReaderClosed(resp)
	if err != nil {
		return response, err
	}

	err = json.NewDecoder(resp.body).Decode(&response)
	return response, err
}
