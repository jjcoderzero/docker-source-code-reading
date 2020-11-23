package client // import "github.com/docker/docker/client"

import (
	"context"
	"io"
	"net/url"

	"github.com/docker/docker/api/types"
)

// ImageLoad从客户端主机加载docker主机中的镜像.
// 由调用者来关闭这个函数返回的ImageLoadResponse中的io.ReadCloser
func (cli *Client) ImageLoad(ctx context.Context, input io.Reader, quiet bool) (types.ImageLoadResponse, error) {
	v := url.Values{}
	v.Set("quiet", "0")
	if quiet {
		v.Set("quiet", "1")
	}
	headers := map[string][]string{"Content-Type": {"application/x-tar"}}
	resp, err := cli.postRaw(ctx, "/images/load", v, input, headers)
	if err != nil {
		return types.ImageLoadResponse{}, err
	}
	return types.ImageLoadResponse{
		Body: resp.body,
		JSON: resp.header.Get("Content-Type") == "application/json",
	}, nil
}
