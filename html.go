package requestsgo

import (
	"io"
	"net/http"
)

// HTML send the requests and parses the response into a Node tree
func (c *Client) HTML(url string, method string, body io.Reader) (Node, error) {
	resp, err := c.Request(url, method, body)
	if err != nil {
		return Node{}, err
	}
	defer resp.Body.Close()

	p, err := ParseHTML(resp.Body)
	if err != nil {
		return Node{}, err
	}

	return p, nil
}

// GetHTML wraps HTML, simplifying sending GET requests
func (c *Client) GetHTML(url string) (Node, error) {
	return c.HTML(url, http.MethodGet, nil)
}
