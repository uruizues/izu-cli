package animepahe

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

const (
	configDir       = ".config/izu-cli"
	cookiesFileName = "cookies.json"
)

// CookieStore manages persistent cookies for Cloudflare bypass
type CookieStore struct {
	cookies []*http.Cookie
	path    string
}

// NewCookieStore creates a new cookie store at the default config location
func NewCookieStore() *CookieStore {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return &CookieStore{
		path: filepath.Join(home, configDir, cookiesFileName),
	}
}

// Load reads cookies from disk
func (cs *CookieStore) Load() error {
	data, err := os.ReadFile(cs.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &cs.cookies)
}

// Save persists cookies to disk
func (cs *CookieStore) Save() error {
	data, err := json.MarshalIndent(cs.cookies, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(cs.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(cs.path, data, 0644)
}

// ApplyToJar loads stored cookies into an HTTP cookie jar
func (cs *CookieStore) ApplyToJar(jar http.CookieJar) {
	if len(cs.cookies) == 0 {
		return
	}
	u, _ := url.Parse(baseURL)
	jar.SetCookies(u, cs.cookies)
}

// UpdateFromJar saves cookies from an HTTP cookie jar
func (cs *CookieStore) UpdateFromJar(jar http.CookieJar) {
	u, _ := url.Parse(baseURL)
	cs.cookies = jar.Cookies(u)
}

// HasCookies returns true if any cookies are stored
func (cs *CookieStore) HasCookies() bool {
	return len(cs.cookies) > 0
}

// Clear removes all stored cookies
func (cs *CookieStore) Clear() error {
	cs.cookies = nil
	return cs.Save()
}

// ImportFromBrowser imports cookies from a Netscape format cookie file
// (typically exported by browser extensions like "Get cookies.txt")
func (cs *CookieStore) ImportFromBrowser(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Parse Netscape cookie format
	lines := splitLines(string(data))
	for _, line := range lines {
		// Skip comments and empty lines
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parts := splitTabs(line)
		if len(parts) < 7 {
			continue
		}

		domain := parts[0]
		// Only import cookies for animepahe domains
		if domain != ".animepahe.com" && domain != "animepahe.com" {
			continue
		}

		cookie := &http.Cookie{
			Name:     parts[5],
			Value:    parts[6],
			Domain:   domain,
			Path:     parts[2],
			Secure:   parts[3] == "TRUE",
			HttpOnly: false,
		}

		cs.cookies = append(cs.cookies, cookie)
	}

	return cs.Save()
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func splitTabs(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\t' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}
