package client // import "docker-19.03/client"

import (
	"context"
	"encoding/json"
	"net/url"

	"docker-19.03/api/types"
	"docker-19.03/api/types/filters"
	"docker-19.03/api/types/swarm"
)

// ConfigList返回配置的列表。
func (cli *Client) ConfigList(ctx context.Context, options types.ConfigListOptions) ([]swarm.Config, error) {
	if err := cli.NewVersionError("1.30", "config list"); err != nil {
		return nil, err
	}
	query := url.Values{}

	if options.Filters.Len() > 0 {
		filterJSON, err := filters.ToJSON(options.Filters)
		if err != nil {
			return nil, err
		}

		query.Set("filters", filterJSON)
	}

	resp, err := cli.get(ctx, "/configs", query, nil)
	defer ensureReaderClosed(resp)
	if err != nil {
		return nil, err
	}

	var configs []swarm.Config
	err = json.NewDecoder(resp.body).Decode(&configs)
	return configs, err
}
