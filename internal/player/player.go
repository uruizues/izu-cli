package player

import (
	"context"
	"time"

	"github.com/izu/izu-cli/internal/provider"
)

type Player interface {
	Play(ctx context.Context, info *provider.StreamInfo, opts PlayOptions) error
	Pause() error
	Stop() error
	Seek(d time.Duration) error
	SetVolume(vol int) error
	GetPosition() (time.Duration, error)
	GetDuration() (time.Duration, error)
	IsRunning() bool
	OnEnd() <-chan struct{}
}

type PlayOptions struct {
	Subtitles []provider.Subtitle
	StartPos  time.Duration
	ExtraArgs []string
}
