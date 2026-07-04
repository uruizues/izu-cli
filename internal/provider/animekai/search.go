package animekai

import (
    "context"
    "strings"

    "github.com/PuerkitoBio/goquery"
    "github.com/izu/izu-cli/internal/provider"
)

func (c *Client) Search(ctx context.Context, query string) ([]provider.SearchResult, error) {
    data, err := c.doSearch(query)
    if err != nil {
        return nil, err
    }

    doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(data)))
    if err != nil {
        return nil, err
    }

    var results []provider.SearchResult
    doc.Find(".anime-list .item").Each(func(i int, s *goquery.Selection) {
        link := s.Find("a")
        href, _ := link.Attr("href")

        parts := strings.Split(href, "/")
        id := ""
        if len(parts) > 0 {
            id = parts[len(parts)-1]
        }

        results = append(results, provider.SearchResult{
            ID:    id,
            Title: strings.TrimSpace(link.Find(".name").Text()),
            Image: s.Find("img").AttrOr("src", ""),
            Type:  strings.TrimSpace(s.Find(".type").Text()),
        })
    })

    return results, nil
}
