package rpc

import (
	"fmt"
	"time"

	"github.com/hugolgst/rich-go/client"
)

type DiscordRPC struct {
	enabled   bool
	clientID  string
	connected bool
}

func NewDiscordRPC(enabled bool, clientID string) *DiscordRPC {
	return &DiscordRPC{
		enabled:  enabled,
		clientID: clientID,
	}
}

func (d *DiscordRPC) Connect() error {
	if !d.enabled || d.clientID == "" {
		return nil
	}

	if err := client.Login(d.clientID); err != nil {
		return err
	}

	d.connected = true
	return nil
}

func (d *DiscordRPC) SetActivity(title, details, state string) error {
	if !d.connected {
		return nil
	}

	return client.SetActivity(client.Activity{
		Details:    details,
		State:      state,
		LargeImage: "anime",
		LargeText:  title,
		Timestamps: &client.Timestamps{
			Start: timePtr(time.Now()),
		},
	})
}

func (d *DiscordRPC) SetWatching(animeTitle string, episode int) error {
	return d.SetActivity(
		animeTitle,
		fmt.Sprintf("Watching %s", animeTitle),
		fmt.Sprintf("Episode %d", episode),
	)
}

func (d *DiscordRPC) ClearActivity() error {
	if !d.connected {
		return nil
	}
	return client.SetActivity(client.Activity{})
}

func (d *DiscordRPC) Disconnect() {
	if d.connected {
		client.Logout()
		d.connected = false
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
