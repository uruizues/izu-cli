package animekai

import (
    "io"
    "net/http"
    "net/url"
    "strings"
    "time"
)

type Client struct {
    baseURL    string
    httpClient *http.Client
    token      string
}

func NewClient(baseURL string) *Client {
    return &Client{
        baseURL: strings.TrimSuffix(baseURL, "/"),
        httpClient: &http.Client{
            Timeout: 15 * time.Second,
        },
    }
}

func (c *Client) doRequest(method, path string, params url.Values) ([]byte, error) {
    u := c.baseURL + path
    if params != nil {
        u += "?" + params.Encode()
    }

    req, err := http.NewRequest(method, u, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
    req.Header.Set("Referer", c.baseURL+"/")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    return io.ReadAll(resp.Body)
}

func (c *Client) doSearch(query string) ([]byte, error) {
    params := url.Values{}
    params.Set("keyword", query)
    params.Set("page", "1")
    return c.doRequest("GET", "/browser", params)
}
