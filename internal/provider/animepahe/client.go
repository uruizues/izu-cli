package animepahe

import (
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

const (
	baseURL = "https://animepahe.com"
	apiURL  = "https://animepahe.com/api.php"
)

type Client struct {
	baseURL    string
	apiURL     string
	httpClient *http.Client
	cookies    *CookieStore
}

func NewClient() *Client {
	jar, _ := cookiejar.New(nil)
	cookieStore := NewCookieStore()
	cookieStore.Load()

	// Apply stored cookies
	cookieStore.ApplyToJar(jar)

	return &Client{
		baseURL: baseURL,
		apiURL:  apiURL,
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: 15 * time.Second,
		},
		cookies: cookieStore,
	}
}

func (c *Client) doRequest(ctx context.Context, u string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Referer", c.baseURL+"/")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Update cookies after request
	c.cookies.UpdateFromJar(c.httpClient.Jar)

	return io.ReadAll(resp.Body)
}

func (c *Client) doSearch(query string) ([]byte, error) {
	u := url.QueryEscape(query)
	return c.doRequest(context.Background(), c.apiURL+"?m=search&q="+u)
}

func (c *Client) GetRelease(id string, page int) ([]byte, error) {
	u := c.apiURL + "?m=release&id=" + id + "&sort=episode_desc&page=" + itoa(page)
	return c.doRequest(context.Background(), u)
}

func (c *Client) GetPlay(id string) ([]byte, error) {
	u := c.apiURL + "?m=play&id=" + id
	return c.doRequest(context.Background(), u)
}

func (c *Client) Name() string { return "animepahe" }

func (c *Client) Close() error {
	c.cookies.UpdateFromJar(c.httpClient.Jar)
	return c.cookies.Save()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
