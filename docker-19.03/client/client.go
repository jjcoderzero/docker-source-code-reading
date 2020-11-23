/*
Package client is a Go client for the Docker Engine API.

For more information about the Engine API, see the documentation:
https://docs.docker.com/engine/reference/api/

Usage

You use the library by creating a client object and calling methods on it. The
client can be created either from environment variables with NewEnvClient, or
configured manually with NewClient.

For example, to list running containers (the equivalent of "docker ps"):

	package main

	import (
		"context"
		"fmt"

		"docker-19.03/api/types"
		"docker-19.03/client"
	)

	func main() {
		cli, err := client.NewClientWithOpts(client.FromEnv)
		if err != nil {
			panic(err)
		}

		containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
		if err != nil {
			panic(err)
		}

		for _, container := range containers {
			fmt.Printf("%s %s\n", container.ID[:10], container.Image)
		}
	}

*/
package client // import "docker-19.03/client"

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"

	"docker-19.03/api"
	"docker-19.03/api/types"
	"docker-19.03/api/types/versions"
	"docker-19.03/go-connections/sockets"
	"github.com/pkg/errors"
)

// ErrRedirect是checkRedirect在请求非get时返回的错误。
var ErrRedirect = errors.New("unexpected redirect in response")

// Client是对docker服务器执行所有操作的API客户机。
type Client struct {
	scheme string // scheme为客户端设置scheme
	host string // host保存要连接到的服务器地址
	proto string // proto持有客户端协议，即unix。
	addr string // addr保存客户端地址。
	basePath string // basePath保存请求的前端路径。
	client *http.Client // client客户端用来发送和接收http请求。
	version string // 要通信的服务器的版本。
	customHTTPHeaders map[string]string // 自定义http头配置的用户。
	manualOverride bool // 当用户设置版本时，manualOverride设置为true。
	negotiateVersion bool // 协商表示客户端在发出请求时是否应该自动协商要使用的API版本。在第一个请求上执行API版本协商，协商后将设置为“true”，以便后续请求不会重新协商
	negotiated bool // negotiation表示进行了API版本协商
}

// CheckRedirect specifies the policy for dealing with redirect responses:
// If the request is non-GET return `ErrRedirect`. Otherwise use the last response.
//
// Go 1.8 changes behavior for HTTP redirects (specifically 301, 307, and 308) in the client .
// The Docker client (and by extension docker API client) can be made to send a request
// like POST /containers//start where what would normally be in the name section of the URL is empty.
// This triggers an HTTP 301 from the daemon.
// In go 1.8 this 301 will be converted to a GET request, and ends up getting a 404 from the daemon.
// This behavior change manifests in the client in that before the 301 was not followed and
// the client did not generate an error, but now results in a message like Error response from daemon: page not found.
func CheckRedirect(req *http.Request, via []*http.Request) error {
	if via[0].Method == http.MethodGet {
		return http.ErrUseLastResponse
	}
	return ErrRedirect
}

// NewClientWithOpts initializes a new API client with default values. It takes functors
// to modify values when creating it, like `NewClientWithOpts(WithVersion(…))`
// It also initializes the custom http headers to add to each request.
//
// It won't send any version information if the version number is empty. It is
// highly recommended that you set a version or your client may break if the
// server is upgraded.
func NewClientWithOpts(ops ...Opt) (*Client, error) {
	client, err := defaultHTTPClient(DefaultDockerHost)
	if err != nil {
		return nil, err
	}
	c := &Client{
		host:    DefaultDockerHost,
		version: api.DefaultVersion,
		client:  client,
		proto:   defaultProto,
		addr:    defaultAddr,
	}

	for _, op := range ops {
		if err := op(c); err != nil {
			return nil, err
		}
	}

	if _, ok := c.client.Transport.(http.RoundTripper); !ok {
		return nil, fmt.Errorf("unable to verify TLS configuration, invalid transport %v", c.client.Transport)
	}
	if c.scheme == "" {
		c.scheme = "http"

		tlsConfig := resolveTLSConfig(c.client.Transport)
		if tlsConfig != nil {
			// TODO(stevvooe): This isn't really the right way to write clients in Go.
			// `NewClient` should probably only take an `*http.Client` and work from there.
			// Unfortunately, the model of having a host-ish/url-thingy as the connection
			// string has us confusing protocol and transport layers. We continue doing
			// this to avoid breaking existing clients but this should be addressed.
			c.scheme = "https"
		}
	}

	return c, nil
}

func defaultHTTPClient(host string) (*http.Client, error) {
	url, err := ParseHostURL(host)
	if err != nil {
		return nil, err
	}
	transport := new(http.Transport)
	sockets.ConfigureTransport(transport, url.Scheme, url.Host)
	return &http.Client{
		Transport:     transport,
		CheckRedirect: CheckRedirect,
	}, nil
}

// Close the transport used by the client
func (cli *Client) Close() error {
	if t, ok := cli.client.Transport.(*http.Transport); ok {
		t.CloseIdleConnections()
	}
	return nil
}

// getAPIPath returns the versioned request path to call the api.
// It appends the query parameters to the path if they are not empty.
func (cli *Client) getAPIPath(ctx context.Context, p string, query url.Values) string {
	var apiPath string
	if cli.negotiateVersion && !cli.negotiated {
		cli.NegotiateAPIVersion(ctx)
	}
	if cli.version != "" {
		v := strings.TrimPrefix(cli.version, "v")
		apiPath = path.Join(cli.basePath, "/v"+v, p)
	} else {
		apiPath = path.Join(cli.basePath, p)
	}
	return (&url.URL{Path: apiPath, RawQuery: query.Encode()}).String()
}

// ClientVersion returns the API version used by this client.
func (cli *Client) ClientVersion() string {
	return cli.version
}

// NegotiateAPIVersion queries the API and updates the version to match the
// API version. Any errors are silently ignored. If a manual override is in place,
// either through the `DOCKER_API_VERSION` environment variable, or if the client
// was initialized with a fixed version (`opts.WithVersion(xx)`), no negotiation
// will be performed.
func (cli *Client) NegotiateAPIVersion(ctx context.Context) {
	if !cli.manualOverride {
		ping, _ := cli.Ping(ctx)
		cli.negotiateAPIVersionPing(ping)
	}
}

// NegotiateAPIVersionPing updates the client version to match the Ping.APIVersion
// if the ping version is less than the default version.  If a manual override is
// in place, either through the `DOCKER_API_VERSION` environment variable, or if
// the client was initialized with a fixed version (`opts.WithVersion(xx)`), no
// negotiation is performed.
func (cli *Client) NegotiateAPIVersionPing(p types.Ping) {
	if !cli.manualOverride {
		cli.negotiateAPIVersionPing(p)
	}
}

// negotiateAPIVersionPing queries the API and updates the version to match the
// API version. Any errors are silently ignored.
func (cli *Client) negotiateAPIVersionPing(p types.Ping) {
	// try the latest version before versioning headers existed
	if p.APIVersion == "" {
		p.APIVersion = "1.24"
	}

	// if the client is not initialized with a version, start with the latest supported version
	if cli.version == "" {
		cli.version = api.DefaultVersion
	}

	// if server version is lower than the client version, downgrade
	if versions.LessThan(p.APIVersion, cli.version) {
		cli.version = p.APIVersion
	}

	// Store the results, so that automatic API version negotiation (if enabled)
	// won't be performed on the next request.
	if cli.negotiateVersion {
		cli.negotiated = true
	}
}

// DaemonHost returns the host address used by the client
func (cli *Client) DaemonHost() string {
	return cli.host
}

// HTTPClient returns a copy of the HTTP client bound to the server
func (cli *Client) HTTPClient() *http.Client {
	return &*cli.client
}

// ParseHostURL parses a url string, validates the string is a host url, and
// returns the parsed URL
func ParseHostURL(host string) (*url.URL, error) {
	protoAddrParts := strings.SplitN(host, "://", 2)
	if len(protoAddrParts) == 1 {
		return nil, fmt.Errorf("unable to parse docker host `%s`", host)
	}

	var basePath string
	proto, addr := protoAddrParts[0], protoAddrParts[1]
	if proto == "tcp" {
		parsed, err := url.Parse("tcp://" + addr)
		if err != nil {
			return nil, err
		}
		addr = parsed.Host
		basePath = parsed.Path
	}
	return &url.URL{
		Scheme: proto,
		Host:   addr,
		Path:   basePath,
	}, nil
}

// CustomHTTPHeaders返回客户端存储的自定义http头。
func (cli *Client) CustomHTTPHeaders() map[string]string {
	m := make(map[string]string)
	for k, v := range cli.customHTTPHeaders {
		m[k] = v
	}
	return m
}

// SetCustomHTTPHeaders that will be set on every HTTP request made by the client.
// Deprecated: use WithHTTPHeaders when creating the client.
func (cli *Client) SetCustomHTTPHeaders(headers map[string]string) {
	cli.customHTTPHeaders = headers
}

// Dialer returns a dialer for a raw stream connection, with HTTP/1.1 header, that can be used for proxying the daemon connection.
// Used by `docker dial-stdio` (docker/cli#889).
func (cli *Client) Dialer() func(context.Context) (net.Conn, error) {
	return func(ctx context.Context) (net.Conn, error) {
		if transport, ok := cli.client.Transport.(*http.Transport); ok {
			if transport.DialContext != nil && transport.TLSClientConfig == nil {
				return transport.DialContext(ctx, cli.proto, cli.addr)
			}
		}
		return fallbackDial(cli.proto, cli.addr, resolveTLSConfig(cli.client.Transport))
	}
}
