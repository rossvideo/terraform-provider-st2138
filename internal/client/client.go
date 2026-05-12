package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"
	"time"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Client carries provider configuration and lazily-dialed connections for both gRPC and HTTP transports.
type Client struct {
	Endpoint   string
	Transport  string
	DevicesDir string

	conn       *grpc.ClientConn
	rpcClient  st2138pb.CatenaServiceClient
	httpClient *HTTPClient
}

// Close releases any underlying connections.
func (c *Client) Close() {
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
		c.rpcClient = nil
	}
	if c.httpClient != nil {
		_ = c.httpClient.Close()
		c.httpClient = nil
	}
}

// Clone returns a shallow copy of the client configuration without any active connection.
// Use per-resource clones to avoid endpoint/connection races across parallel resources.
func (c *Client) Clone() *Client {
	return &Client{
		Endpoint:   c.Endpoint,
		Transport:  c.Transport,
		DevicesDir: c.DevicesDir,
		conn:       nil,
		rpcClient:  nil,
		httpClient: nil,
	}
}

// SetEndpoint updates the client's endpoint. If the endpoint changes while a
// connection is open, the existing connection is closed so the next operation
// will re-dial the new target.
func (c *Client) SetEndpoint(ep string) {
	if ep == "" {
		return
	}
	if c.Endpoint != ep {
		c.Endpoint = ep
		c.Close()
	}
}

// ensureConn dials the gRPC endpoint and initializes the RPC client.
func (c *Client) ensureConn(ctx context.Context) error {
	if c.Transport != "grpc" {
		return fmt.Errorf("transport %q is not grpc", c.Transport)
	}
	if c.conn != nil && c.rpcClient != nil {
		return nil
	}
	// Dial timeout for establishing connection
	dctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	target := c.Endpoint
	var opts []grpc.DialOption
	// If endpoint includes an explicit scheme (://), use it to decide TLS
	// e.g., https://host:port or grpcs://host:port. Otherwise treat as host[:port].
	if strings.Contains(target, "://") {
		if u, err := url.Parse(target); err == nil {
			host := u.Host
			if host == "" {
				host = u.Path
			}
			if u.Scheme == "https" || u.Scheme == "grpcs" {
				serverName := host
				if hp := strings.Split(host, ":"); len(hp) > 0 {
					serverName = hp[0]
				}
				tlsCfg := &tls.Config{ServerName: serverName}
				opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
			} else {
				opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
			}
			target = host
		} else {
			opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}
	} else {
		// No scheme provided: default to insecure and use target as-is
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.DialContext(dctx, target, opts...)
	if err != nil {
		return err
	}
	c.conn = conn
	c.rpcClient = st2138pb.NewCatenaServiceClient(conn)
	return nil
}

// ensureHTTPConn establishes an HTTP client connection.
func (c *Client) ensureHTTPConn(ctx context.Context) error {
	if c.Transport != "http" && c.Transport != "https" && c.Transport != "rest" {
		return fmt.Errorf("transport %q is not http", c.Transport)
	}
	if c.httpClient != nil {
		return nil
	}

	// Normalize transport: map "rest" to "http" or "https" based on endpoint
	transport := c.Transport
	if transport == "rest" {
		if strings.Contains(c.Endpoint, "https://") {
			transport = "https"
		} else {
			transport = "http"
		}
	}

	// If endpoint doesn't have scheme, add it
	endpoint := c.Endpoint
	if !strings.Contains(endpoint, "://") {
		endpoint = transport + "://" + endpoint
	}

	httpClient, err := NewHTTPClient(endpoint)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}

	c.httpClient = httpClient
	return nil
}
