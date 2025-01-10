package requestsgo

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// Config for creating a Client
type Config struct {
	UserAgent  string
	Referer    string
	CreateDirs bool
}

// Client manages all requests
type Client struct {
	ctx    context.Context
	client *http.Client
	cfg    Config
}

// NewClient creates a new instance of a Client
func NewClient(ctx context.Context, cfg Config) *Client {
	return &Client{ctx: ctx, client: &http.Client{}, cfg: cfg}
}

// Request sends a HTTP request, returning the *http.Response
func (c *Client) Request(url string, method string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(c.ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}

	if len(c.cfg.UserAgent) > 0 {
		req.Header.Set("User-Agent", c.cfg.UserAgent)
	}

	if len(c.cfg.Referer) > 0 {
		req.Header.Set("Referer", c.cfg.Referer)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request failed: %w", err)
	}

	return resp, nil
}
