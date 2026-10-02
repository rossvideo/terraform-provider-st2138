package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const catenaRESTPrefix = "/st2138-api/v1"

var catenaRESTHTTPClient = &http.Client{Timeout: 30 * time.Second}

func (c *Client) usesREST() bool {
	return strings.EqualFold(c.Transport, "rest")
}

func (c *Client) restRequest(ctx context.Context, method, route string, request proto.Message, headers map[string]string, response proto.Message) ([]byte, error) {
	baseURL := c.Endpoint
	if !strings.Contains(baseURL, "://") {
		baseURL = "http://" + baseURL
	}
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse REST endpoint %q: %w", c.Endpoint, err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("unsupported REST endpoint scheme %q", parsedURL.Scheme)
	}
	if parsedURL.Host == "" {
		return nil, fmt.Errorf("REST endpoint %q has no host", c.Endpoint)
	}

	routePath, query, _ := strings.Cut(route, "?")
	parsedURL.Path = strings.TrimRight(parsedURL.Path, "/") + catenaRESTPrefix + routePath
	parsedURL.RawQuery = query

	var body []byte
	if request != nil {
		body, err = (protojson.MarshalOptions{UseProtoNames: true}).Marshal(request)
		if err != nil {
			return nil, fmt.Errorf("marshal REST request: %w", err)
		}
	}
	requestMessage, err := http.NewRequestWithContext(ctx, method, parsedURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create REST request: %w", err)
	}
	requestMessage.Header.Set("Accept", "application/json")
	if request != nil {
		requestMessage.Header.Set("Content-Type", "application/json")
	}
	for name, value := range headers {
		requestMessage.Header.Set(name, value)
	}

	httpResponse, err := catenaRESTHTTPClient.Do(requestMessage)
	if err != nil {
		return nil, fmt.Errorf("REST %s %s: %w", method, parsedURL.Redacted(), err)
	}
	defer httpResponse.Body.Close()
	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("read REST response: %w", err)
	}
	if httpResponse.StatusCode < http.StatusOK || httpResponse.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("REST %s %s returned HTTP %d: %s", method, parsedURL.Redacted(), httpResponse.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	if response != nil && len(responseBody) > 0 {
		if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(responseBody, response); err != nil {
			return nil, fmt.Errorf("decode REST response: %w", err)
		}
	}
	if response != nil && len(responseBody) == 0 {
		return nil, fmt.Errorf("REST %s %s returned an empty response", method, parsedURL.Redacted())
	}
	return responseBody, nil
}
