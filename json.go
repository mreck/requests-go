package requestsgo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// JSON send the requests and parses the response into a the dst
func (c *Client) JSON(url string, method string, body io.Reader, dst interface{}) error {
	resp, err := c.Request(url, method, body)
	if err != nil {
		return fmt.Errorf("executing request failed: %w", err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(dst)
	if err != nil {
		return fmt.Errorf("decoding JSON failed: %w", err)
	}

	return nil
}

// GetJSON wraps JSON, simplifying sending GET requests
func (c *Client) GetJSON(url string, dst interface{}) error {
	return c.JSON(url, http.MethodGet, nil, dst)
}
