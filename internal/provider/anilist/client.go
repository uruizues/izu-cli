package anilist

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

const APIURL = "https://graphql.anilist.co"

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

func (c *Client) doQuery(query string, variables map[string]interface{}) ([]byte, error) {
	body := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", APIURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "izu-cli/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *Client) Name() string  { return "anilist" }
func (c *Client) Close() error { return nil }
