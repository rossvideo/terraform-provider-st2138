package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
	"google.golang.org/protobuf/encoding/protojson"
)

// HTTPClient provides REST API access to Catena devices via HTTP transport.
type HTTPClient struct {
	endpoint   string
	httpClient *http.Client
	baseURL    string
	scheme     string
}

// NewHTTPClient creates a new HTTP client for the given endpoint.
// Endpoint should be in the form "host:port" or "scheme://host:port".
func NewHTTPClient(endpoint string) (*HTTPClient, error) {
	hc := &HTTPClient{
		endpoint: endpoint,
	}

	// Parse endpoint and determine scheme
	if strings.Contains(endpoint, "://") {
		parts := strings.Split(endpoint, "://")
		scheme := parts[0]
		hostport := parts[1]
		hc.scheme = scheme
		hc.baseURL = fmt.Sprintf("%s://%s/st2138-api/v1", scheme, hostport)
	} else {
		// Default to http for plain host:port
		hc.scheme = "http"
		hc.baseURL = fmt.Sprintf("http://%s/st2138-api/v1", endpoint)
	}

	// Configure TLS if needed
	tlsConfig := &tls.Config{InsecureSkipVerify: false}
	if hc.scheme == "https" || hc.scheme == "grpcs" {
		host := strings.Split(endpoint, ":")[0]
		if strings.Contains(endpoint, "://") {
			host = strings.Split(strings.Split(endpoint, "://")[1], ":")[0]
		}
		tlsConfig.ServerName = host
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Use custom timeout for connections
			d := net.Dialer{Timeout: 10 * time.Second}
			return d.DialContext(ctx, network, addr)
		},
	}

	hc.httpClient = &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return hc, nil
}

// Close releases HTTP client resources.
func (hc *HTTPClient) Close() error {
	if hc.httpClient != nil {
		hc.httpClient.CloseIdleConnections()
	}
	return nil
}

// DeviceRequest streams device information via HTTP.
// Returns a stream-like interface for compatibility with gRPC usage.
func (hc *HTTPClient) DeviceRequest(ctx context.Context, slot uint32) (*HTTPStream, error) {
	url := fmt.Sprintf("%s/%d", hc.baseURL, slot)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("device request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return &HTTPStream{
		reader:  resp.Body,
		decoder: json.NewDecoder(resp.Body),
	}, nil
}

// GetValue fetches a parameter value via REST API.
func (hc *HTTPClient) GetValue(ctx context.Context, slot uint32, oid string) (*st2138pb.Value, error) {
	if !strings.HasPrefix(oid, "/") {
		oid = "/" + oid
	}

	// Encode OID for URL path
	encodedOid := strings.ReplaceAll(strings.TrimPrefix(oid, "/"), "/", ":")

	url := fmt.Sprintf("%s/%d/value/%s", hc.baseURL, slot, encodedOid)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get value failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Value *st2138pb.Value `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Value, nil
}

// SetValue updates a parameter value via REST API.
func (hc *HTTPClient) SetValue(ctx context.Context, slot uint32, oid string, value *st2138pb.Value) error {
	if !strings.HasPrefix(oid, "/") {
		oid = "/" + oid
	}

	encodedOid := strings.ReplaceAll(strings.TrimPrefix(oid, "/"), "/", ":")
	url := fmt.Sprintf("%s/%d/value/%s", hc.baseURL, slot, encodedOid)

	payload := map[string]interface{}{
		"value": value,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("set value failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// ExecuteCommand executes a device command via REST API.
func (hc *HTTPClient) ExecuteCommand(ctx context.Context, slot uint32, oid string, value *st2138pb.Value, respond bool) (*HTTPStream, error) {
	if !strings.HasPrefix(oid, "/") {
		oid = "/" + oid
	}

	encodedOid := strings.ReplaceAll(strings.TrimPrefix(oid, "/"), "/", ":")
	url := fmt.Sprintf("%s/%d/command/%s/stream", hc.baseURL, slot, encodedOid)

	payload := map[string]interface{}{
		"value":   value,
		"respond": respond,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("execute command failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return &HTTPStream{
		reader:  resp.Body,
		decoder: json.NewDecoder(resp.Body),
	}, nil
}

// GetParam fetches parameter descriptor via REST API.
func (hc *HTTPClient) GetParam(ctx context.Context, slot uint32, oid string) (*st2138pb.Param, error) {
	if !strings.HasPrefix(oid, "/") {
		oid = "/" + oid
	}

	encodedOid := strings.ReplaceAll(strings.TrimPrefix(oid, "/"), "/", ":")
	url := fmt.Sprintf("%s/%d/param-info/%s", hc.baseURL, slot, encodedOid)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get param failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Param *st2138pb.Param `json:"param"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Param, nil
}

// HTTPStream provides a stream-like interface for HTTP responses.
type HTTPStream struct {
	reader  io.ReadCloser
	decoder *json.Decoder
	current interface{}
}

// Recv reads the next message from the stream.
// For now, returns raw bytes as st2138pb.DeviceComponent for device streams.
func (hs *HTTPStream) Recv() (*st2138pb.DeviceComponent, error) {
	if hs.decoder == nil {
		return nil, io.EOF
	}

	var raw json.RawMessage
	if err := hs.decoder.Decode(&raw); err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, err
	}

	// First try full DeviceComponent shape using protobuf JSON decoding.
	var component st2138pb.DeviceComponent
	if err := protojson.Unmarshal(raw, &component); err == nil {
		if component.GetDevice() != nil || component.GetParam() != nil || component.GetCommand() != nil {
			return &component, nil
		}
	}

	// Then try plain device object shape using protobuf JSON decoding.
	var dev st2138pb.Device
	if err := protojson.Unmarshal(raw, &dev); err == nil {
		return &st2138pb.DeviceComponent{
			Kind: &st2138pb.DeviceComponent_Device{Device: &dev},
		}, nil
	}

	// Finally try wrapper object shape: {"device": {...}}
	var wrapped struct {
		Device json.RawMessage `json:"device"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil && len(wrapped.Device) > 0 {
		var wrappedDev st2138pb.Device
		if uerr := protojson.Unmarshal(wrapped.Device, &wrappedDev); uerr == nil {
			return &st2138pb.DeviceComponent{
				Kind: &st2138pb.DeviceComponent_Device{Device: &wrappedDev},
			}, nil
		}
	}

	return nil, fmt.Errorf("unsupported device payload format")
}

// RecvCommandResponse reads the next command response from the stream.
func (hs *HTTPStream) RecvCommandResponse() (*st2138pb.CommandResponse, error) {
	if hs.decoder == nil {
		return nil, io.EOF
	}

	var response st2138pb.CommandResponse
	err := hs.decoder.Decode(&response)
	if err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, err
	}

	return &response, nil
}

// Close closes the stream.
func (hs *HTTPStream) Close() error {
	if hs.reader != nil {
		return hs.reader.Close()
	}
	return nil
}
