package player

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type DownloadProgress struct {
	Percent float64
	Speed   string
	ETA     string
	State   string // downloading, muxing, done, error
	Error   error
}

type Downloader struct {
	outputDir string
	ffmpeg    string
}

func NewDownloader(outputDir string) *Downloader {
	ffmpeg, _ := exec.LookPath("ffmpeg")

	return &Downloader{
		outputDir: outputDir,
		ffmpeg:    ffmpeg,
	}
}

func (d *Downloader) Download(ctx context.Context, m3u8URL, outputName string, progress chan<- DownloadProgress) error {
	os.MkdirAll(d.outputDir, 0755)

	outputPath := filepath.Join(d.outputDir, outputName+".mp4")

	if d.ffmpeg != "" {
		return d.downloadWithFFmpeg(ctx, m3u8URL, outputPath, progress)
	}

	return d.downloadManual(ctx, m3u8URL, outputPath, progress)
}

func (d *Downloader) downloadWithFFmpeg(ctx context.Context, m3u8URL, outputPath string, progress chan<- DownloadProgress) error {
	progress <- DownloadProgress{State: "downloading", Percent: 0}

	cmd := exec.CommandContext(ctx, d.ffmpeg,
		"-i", m3u8URL,
		"-c", "copy",
		"-bsf:a", "aac_adtstoasc",
		outputPath,
	)

	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		progress <- DownloadProgress{State: "error", Error: err}
		return err
	}

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "time=") {
			progress <- DownloadProgress{State: "downloading", Percent: 50}
		}
	}

	if err := cmd.Wait(); err != nil {
		progress <- DownloadProgress{State: "error", Error: err}
		return err
	}

	progress <- DownloadProgress{State: "done", Percent: 100}
	return nil
}

func (d *Downloader) downloadManual(ctx context.Context, m3u8URL, outputPath string, progress chan<- DownloadProgress) error {
	progress <- DownloadProgress{State: "downloading", Percent: 0}

	resp, err := http.Get(m3u8URL)
	if err != nil {
		progress <- DownloadProgress{State: "error", Error: err}
		return err
	}
	defer resp.Body.Close()

	var segments []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			segments = append(segments, line)
		}
	}

	if len(segments) == 0 {
		err := fmt.Errorf("no segments found in m3u8")
		progress <- DownloadProgress{State: "error", Error: err}
		return err
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		progress <- DownloadProgress{State: "error", Error: err}
		return err
	}
	defer outFile.Close()

	var mu sync.Mutex
	total := len(segments)

	for i, seg := range segments {
		select {
		case <-ctx.Done():
			progress <- DownloadProgress{State: "error", Error: ctx.Err()}
			return ctx.Err()
		default:
		}

		if !strings.HasPrefix(seg, "http") {
			lastSlash := strings.LastIndex(m3u8URL, "/")
			if lastSlash > 0 {
				seg = m3u8URL[:lastSlash+1] + seg
			}
		}

		segResp, err := http.Get(seg)
		if err != nil {
			continue
		}

		mu.Lock()
		io.Copy(outFile, segResp.Body)
		mu.Unlock()
		segResp.Body.Close()

		percent := float64(i+1) / float64(total) * 100
		progress <- DownloadProgress{
			State:   "downloading",
			Percent: percent,
		}
	}

	progress <- DownloadProgress{State: "done", Percent: 100}
	return nil
}
