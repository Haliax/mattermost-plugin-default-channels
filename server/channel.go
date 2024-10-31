package main

import (
	"fmt"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// UserHasLeftChannel is invoked after the membership has been removed from the database. If
// actor is not nil, the user was removed from the channel by the actor.
func (p *Plugin) UserHasLeftChannel(c *plugin.Context, channelMember *model.ChannelMember, actor *model.User) {
	// ToDo: Is a disabled check necessary?

	user, err := p.API.GetUser(channelMember.UserId)
	if err != nil {
		p.API.LogError(
			"Failed to query user",
			"user_id", channelMember.UserId,
			"error", err.Error(),
		)
		return
	}

	channel, err := p.API.GetChannel(channelMember.ChannelId)
	if err != nil {
		p.API.LogError(
			"Failed to query channel",
			"channel_id", channelMember.ChannelId,
			"error", err.Error(),
		)
		return
	}

	msg := fmt.Sprintf("UserHasLeftChannel: @%s, ~%s", user.Username, channel.Name)
	p.API.LogInfo(msg)

	// ToDo: Check if the channel was a default channel and add the user again.
}
