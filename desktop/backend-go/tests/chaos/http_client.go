package chaos

import (
	"bytes"
	"net/http"
	"time"
)

// httpClient is a simple HTTP client wrapper for health checks
type httpClient struct {
	timeout time.Duration
	client  *http.Client
}

// NewHTTPClient creates a new HTTP client with the specified timeout
func NewHTTPClient(timeout time.Duration) *httpClient {
	return &httpClient{
		timeout: timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Get performs an HTTP GET request
func (c *httpClient) Get(url string) (*http.Response, error) {
	return c.client.Get(url)
}

// Post performs an HTTP POST request
func (c *httpClient) Post(url, contentType string, body []byte) (*http.Response, error) {
	return c.client.Post(url, contentType, bytes.NewReader(body))
}

// Head performs an HTTP HEAD request
func (c *httpClient) Head(url string) (*http.Response, error) {
	return c.client.Head(url)
}
