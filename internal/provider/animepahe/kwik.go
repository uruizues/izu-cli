package animepahe

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/dop251/goja"
)

const kwikBaseURL = "https://kwik.cx"

// extractStreamURL fetches the kwik.cx page and extracts the real m3u8 stream URL
func extractStreamURL(ctx context.Context, client *http.Client, kwikPath string) (string, error) {
	kwikURL := kwikBaseURL + kwikPath

	req, err := http.NewRequestWithContext(ctx, "GET", kwikURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://animepahe.com/")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	html, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return decodeKwikJS(string(html))
}

// decodeKwikJS extracts and executes the obfuscated JS to get the stream URL
func decodeKwikJS(html string) (string, error) {
	re := regexp.MustCompile(`eval\(function\(p,a,c,k,e,d\)[^)]*\)`)
	match := re.FindString(html)
	if match == "" {
		re2 := regexp.MustCompile(`eval\(function\(p,a,c,k,e,d\).*?</script>`)
		match = re2.FindString(html)
		if match == "" {
			return "", fmt.Errorf("no JS payload found in kwik page")
		}
		match = strings.TrimSuffix(match, "</script>")
	}

	vm := goja.New()
	setupMockDOM(vm)

	var capturedURL string
	var logs []string

	document := vm.NewObject()
	document.Set("createElement", func(tag string) goja.Value {
		el := vm.NewObject()
		el.Set("setAttribute", func(key, value string) {})
		el.Set("innerHTML", goja.Undefined())
		el.Set("defineProperty", func(name string, descriptor goja.Value) goja.Value {
			return goja.Undefined()
		})
		return vm.ToValue(el)
	})
	vm.Set("document", document)

	console := vm.NewObject()
	console.Set("log", func(args ...goja.Value) {
		for _, arg := range args {
			logs = append(logs, arg.String())
		}
	})
	vm.Set("console", console)

	window := vm.NewObject()
	vm.Set("window", window)

	_, err := vm.RunString(match)
	if err != nil {
		return "", fmt.Errorf("JS execution failed: %w", err)
	}

	for _, log := range logs {
		if strings.Contains(log, "m3u8") || strings.Contains(log, ".mp4") || strings.Contains(log, "https://") {
			capturedURL = log
			break
		}
	}

	if capturedURL == "" {
		capturedURL = extractURLFromPacked(match)
	}

	if capturedURL == "" {
		return "", fmt.Errorf("could not extract stream URL from kwik page")
	}

	return capturedURL, nil
}

func setupMockDOM(vm *goja.Runtime) {
	navigator := vm.NewObject()
	navigator.Set("userAgent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	vm.Set("navigator", navigator)

	localStorage := vm.NewObject()
	localStorage.Set("getItem", func(key string) goja.Value { return goja.Null() })
	localStorage.Set("setItem", func(key, value string) {})
	vm.Set("localStorage", localStorage)

	vm.Set("atob", func(encoded string) string {
		return encoded
	})
}

func extractURLFromPacked(js string) string {
	patterns := []string{
		`source\s+src\s*[:=]\s*["']([^"']+m3u8[^"']*)["']`,
		`file\s*[:=]\s*["']([^"']+\.m3u8[^"']*)["']`,
		`(?:url|src|source)\s*[:=]\s*["'](https?://[^"']+\.m3u8[^"']*)["']`,
		`(https?://[^"'\s]+/[^"'\s]+\.m3u8[^"'\s]*)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(js)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}
