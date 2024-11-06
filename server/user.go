package main

import (
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// UserHasLoggedIn is invoked after a user has logged in.
func (p *Plugin) UserHasLoggedIn(c *plugin.Context, user *model.User) {
	p.addToAllDefaultChannels(user)
}

// UserHasBeenCreated is invoked when a new user is created.
func (p *Plugin) UserHasBeenCreated(c *plugin.Context, user *model.User) {
	// Not sure whether this hook is useful or not.
}
