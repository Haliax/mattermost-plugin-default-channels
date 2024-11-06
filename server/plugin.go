package main

import (
	"sync"

	"github.com/mattermost/mattermost/server/public/plugin"
)

// ToDo: Periodically add users to channel.

// ToDo: Prevent users from leaving the channel.
// ToDo: Create 'add all to channel' command.

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	// BotId of the created bot account.
	botID string
}

// OnActivate is invoked when the plugin is activated.
//
// This demo implementation logs a message to the demo channel whenever the plugin is activated.
// It also creates a demo bot account
func (p *Plugin) OnActivate() error {
	p.addAllUsersToDefaultChannels()
	return nil
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
