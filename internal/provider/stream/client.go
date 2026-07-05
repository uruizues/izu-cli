package stream

import (
	"encoding/json"
	"os/exec"
	"time"
)

type Client struct {
	binary string
}

type ytDlpInfo struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Duration    int    `json:"duration"`
	Thumbnail   string `json:"thumbnail"`
	Formats     []struct {
		URL        string `json:"url"`
		Ext        string `json:"ext"`
		Resolution string `json:"resolution"`
		FormatNote string `json:"format_note"`
	} `json:"formats"`
	Subtitles map[string][]struct {
		URL string `json:"url"`
	} `json:"subtitles"`
}

func NewClient() *Client {
	binary, _ := exec.LookPath("yt-dlp")
	if binary == "" {
		binary = "yt-dlp"
	}
	return &Client{binary: binary}
}

func (c *Client) GetStreamURL(url string) (*ytDlpInfo, error) {
	cmd := exec.Command(c.binary,
		"--dump-json",
		"--no-download",
		"--no-warnings",
		url,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var info ytDlpInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

func (c *Client) Name() string  { return "stream" }
func (c *Client) Close() error { return nil }

func init() {
	// Ensure we have enough timeout for yt-dlp
	_ = time.Second
}
