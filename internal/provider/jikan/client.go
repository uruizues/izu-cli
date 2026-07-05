package jikan

import (
	"io"
	"net/http"
	"time"
)

const BaseURL = "https://api.jikan.moe/v4"

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) doRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "izu-cli/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Rate limit: Jikan allows 3 requests per second
	time.Sleep(350 * time.Millisecond)

	return io.ReadAll(resp.Body)
}

func (c *Client) Name() string  { return "jikan" }
func (c *Client) Close() error { return nil }
