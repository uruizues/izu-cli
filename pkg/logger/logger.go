package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var log zerolog.Logger

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	log = zerolog.New(output).With().Timestamp().Logger()
}

func Info() *zerolog.Event  { return log.Info() }
func Error() *zerolog.Event { return log.Error() }
func Debug() *zerolog.Event { return log.Debug() }
func Warn() *zerolog.Event  { return log.Warn() }
