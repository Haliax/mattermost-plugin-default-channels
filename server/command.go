package main

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

const (
	commandTriggerHooks = "default_channels"
)

func (p *Plugin) registerCommands() error {

	if err := p.API.RegisterCommand(&model.Command{
		Trigger:          commandTriggerHooks,
		AutoComplete:     true,
		AutoCompleteHint: "",
		AutoCompleteDesc: "Ensures all users and guests are added to their respecting default channels..",
	}); err != nil {
		return errors.Wrapf(err, "failed to register %s command", commandTriggerHooks)
	}

	return nil
}

// ExecuteCommand executes a command that has been previously registered via the RegisterCommand API.
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	trigger := strings.TrimPrefix(strings.Fields(args.Command)[0], "/")
	switch trigger {
	case commandTriggerHooks:
		return p.executeCommandHooks(args), nil

	default:
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Unknown command: " + args.Command),
		}, nil
	}
}

func (p *Plugin) executeCommandHooks(args *model.CommandArgs) *model.CommandResponse {
	user, err := p.API.GetUser(args.UserId)

	if err != nil || !user.IsSystemAdmin() {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "You must be a system administrator to run this command.",
		}
	}

	p.addAllUsersToDefaultChannels()

	_ = p.API.SendEphemeralPost(args.UserId, &model.Post{
		ChannelId: args.ChannelId,
		Message:   "All default channels have been synced.",
		Props: model.StringInterface{
			"type": "default_channels_plugin_ephemeral",
		},
	})

	return &model.CommandResponse{}
}
