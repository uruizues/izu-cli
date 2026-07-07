package proxy

import (
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
)

var (
	server  *http.Server
	port    int
	mu      sync.Mutex
	counter atomic.Int32
)

func Start(referer, origin string) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	if server != nil {
		return port, nil
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}

	port = ln.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		counter.Add(1)

		targetURL := r.URL.Query().Get("url")
		if targetURL == "" {
			http.Error(w, "missing url param", 400)
			return
		}

		proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		proxyReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		if referer != "" {
			proxyReq.Header.Set("Referer", referer)
		}
		if origin != "" {
			proxyReq.Header.Set("Origin", origin)
		}

		resp, err := http.DefaultClient.Do(proxyReq)
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		defer resp.Body.Close()

		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	server = &http.Server{Handler: mux}
	go server.Serve(ln)

	return port, nil
}

func Stop() {
	mu.Lock()
	defer mu.Unlock()

	if server != nil {
		server.Close()
		server = nil
	}
	port = 0
}

func ProxyURL(originalURL string) string {
	if port == 0 {
		return originalURL
	}
	return "http://127.0.0.1:" + strconv.Itoa(port) + "/?url=" + url.QueryEscape(originalURL)
}

func Hits() int64 {
	return int64(counter.Load())
}
