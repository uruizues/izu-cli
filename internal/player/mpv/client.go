package mpv

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/izu/izu-cli/internal/player"
	"github.com/izu/izu-cli/internal/provider"
)

type Client struct {
	binary     string
	args       []string
	socketPath string
	cmd        *exec.Cmd
	conn       net.Conn
	done       chan struct{}
}

func New(binary string, args []string, socketPath string) *Client {
	return &Client{
		binary:     binary,
		args:       args,
		socketPath: socketPath,
		done:       make(chan struct{}),
	}
}

func (c *Client) Play(ctx context.Context, info *provider.StreamInfo, opts player.PlayOptions) error {
	args := append(c.args, "--input-ipc-server="+c.socketPath)

	if info.Referer != "" {
		args = append(args, "--http-header-fields=Referer: "+info.Referer)
	}

	for k, v := range info.Headers {
		args = append(args, "--http-header-fields="+k+": "+v)
	}

	for _, sub := range opts.Subtitles {
		args = append(args, "--sub-file="+sub.URL)
	}

	if opts.StartPos > 0 {
		args = append(args, fmt.Sprintf("--start=%s", opts.StartPos))
	}

	args = append(args, opts.ExtraArgs...)
	args = append(args, info.URL)

	c.cmd = exec.CommandContext(ctx, c.binary, args...)
	c.cmd.Stdout = os.Stdout
	c.cmd.Stderr = os.Stderr

	if err := c.cmd.Start(); err != nil {
		return err
	}

	go func() {
		c.cmd.Wait()
		close(c.done)
	}()

	time.Sleep(500 * time.Millisecond)

	return c.connect()
}

func (c *Client) connect() error {
	var err error
	for i := 0; i < 10; i++ {
		c.conn, err = net.Dial("unix", c.socketPath)
		if err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return err
}

func (c *Client) sendCommand(cmd string) error {
	if c.conn == nil {
		return nil
	}
	_, err := c.conn.Write([]byte(cmd + "\n"))
	return err
}

func (c *Client) Pause() error                       { return c.sendCommand("cycle pause") }
func (c *Client) Stop() error                        { return c.sendCommand("quit") }
func (c *Client) Seek(d time.Duration) error         { return c.sendCommand("seek " + d.String()) }
func (c *Client) SetVolume(vol int) error            { return c.sendCommand(fmt.Sprintf("set volume %d", vol)) }
func (c *Client) GetPosition() (time.Duration, error) { return 0, nil }
func (c *Client) GetDuration() (time.Duration, error) { return 0, nil }
func (c *Client) IsRunning() bool                    { return c.cmd != nil && c.cmd.Process != nil }
func (c *Client) OnEnd() <-chan struct{}              { return c.done }
