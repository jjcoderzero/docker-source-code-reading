package client // import "docker-19.03/client"

import (
	"context"
	"encoding/json"
	"net/url"

	"docker-19.03/api/types/container"
)

// ContainerDiff shows differences in a container filesystem since it was started.
func (cli *Client) ContainerDiff(ctx context.Context, containerID string) ([]container.ContainerChangeResponseItem, error) {
	var changes []container.ContainerChangeResponseItem

	serverResp, err := cli.get(ctx, "/containers/"+containerID+"/changes", url.Values{}, nil)
	defer ensureReaderClosed(serverResp)
	if err != nil {
		return changes, err
	}

	err = json.NewDecoder(serverResp.body).Decode(&changes)
	return changes, err
}
