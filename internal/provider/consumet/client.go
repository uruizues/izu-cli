package consumet

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/izu/izu-cli/internal/provider"
)

const DefaultBaseURL = "https://api.consumet.org"

var DefaultProviders = []string{
	"gogoanime",
	"zoro",
	"animepahe",
	"9anime",
	"animefox",
	"enime",
	"crunchyroll",
	"bilibili",
	"marin",
	"animesaturn",
}

type Client struct {
	provider   string
	providers  []string
	baseURL    string
	httpClient *http.Client
}

func NewClient(provider string) *Client {
	return &Client{
		provider:  provider,
		providers: []string{provider},
		baseURL:   DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func NewClientWithBaseURL(provider, baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		provider:  provider,
		providers: []string{provider},
		baseURL:   baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func NewMultiProviderClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		provider:  "multi",
		providers: DefaultProviders,
		baseURL:   baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (c *Client) Name() string {
	if c.provider == "multi" {
		return "consumet-all"
	}
	return "consumet-" + c.provider
}

func (c *Client) Close() error { return nil }

func (c *Client) searchSingleProvider(query string, prov string) []provider.SearchResult {
	results := make(chan provider.SearchResult, 50)

	go func() {
		defer close(results)
		u := fmt.Sprintf("%s/anime/%s/%s?page=1", c.baseURL, prov, query)
		data, err := c.doRequest(u)
		if err != nil {
			return
		}

		var resp searchResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return
		}

		for _, r := range resp.Results {
			results <- provider.SearchResult{
				ID:       r.ID,
				Title:    r.Title,
				Image:    r.Image,
				Type:     r.SubOrDub,
				Episodes: 0,
				Status:   prov,
			}
		}
	}()

	var collected []provider.SearchResult
	for r := range results {
		collected = append(collected, r)
	}
	return collected
}

func (c *Client) searchAllProviders(ctx context.Context, query string) ([]provider.SearchResult, error) {
	var mu sync.Mutex
	var wg sync.WaitGroup
	var allResults []provider.SearchResult

	for _, prov := range c.providers {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			results := c.searchSingleProvider(query, p)
			mu.Lock()
			allResults = append(allResults, results...)
			mu.Unlock()
		}(prov)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return allResults, ctx.Err()
	case <-done:
	}

	return allResults, nil
}
