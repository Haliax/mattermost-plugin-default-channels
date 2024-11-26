package main

import (
	"github.com/mattermost/mattermost/server/public/model"
)

func (p *Plugin) isDefaultChannelForUser(channel *model.Channel, user *model.User) bool {
	configuration := p.getConfiguration()
	if configuration.memberChannelIDs == nil {
		return false
	}

	if user.IsGuest() {
		return contains(configuration.guestChannelIDs[channel.TeamId], channel.Id)
	}

	return contains(configuration.memberChannelIDs[channel.TeamId], channel.Id)
}

func (p *Plugin) addAllUsersToDefaultChannels() {
	// Get all users of mattermost instance via API without pag
	users, err := p.API.GetUsers(&model.UserGetOptions{
		Page:    0,
		PerPage: 99999,
		Active:  true,
	})
	if err != nil {
		p.API.LogError("Failed to get users", "error", err.Error())
		return
	}

	for _, user := range users {
		p.addToAllDefaultChannels(user, false)
	}
}

func (p *Plugin) addToAllDefaultChannels(user *model.User, silent bool) {
	configuration := p.getConfiguration()

	if user.IsGuest() {
		for _, channelIDs := range configuration.guestChannelIDs {
			for _, channelID := range channelIDs {
				channel, err := p.API.GetChannel(channelID)
				if err != nil {
					continue
				}

				p.addUserToDefaultChannel(channel, user, silent)
			}
		}
		return
	}

	for _, channelIDs := range configuration.memberChannelIDs {
		for _, channelID := range channelIDs {
			channel, err := p.API.GetChannel(channelID)
			if err != nil {
				continue
			}

			p.addUserToDefaultChannel(channel, user, silent)
		}
	}
}

func (p *Plugin) addUserToDefaultChannel(channel *model.Channel, user *model.User, silent bool) bool {
	team, err := p.API.GetTeam(channel.TeamId)
	if err != nil {
		return false
	}

	p.ensureTeamMembership(team.Id, user.Id)

	if silent {
		_, err = p.API.AddChannelMember(channel.Id, user.Id)
		return err != nil
	}

	_, err = p.API.AddUserToChannel(channel.Id, user.Id, p.botID)
	return err != nil
}

func (p *Plugin) ensureTeamMembership(teamID string, userID string) bool {
	member, teamError := p.API.GetTeamMember(teamID, userID)

	if teamError != nil || member == nil || member.DeleteAt != 0 {
		config := p.getConfiguration()
		if !config.AddToTeam {
			p.API.LogDebug("Skipping adding user to team", "team_id", teamID, "user_id", userID)
			return false
		}

		p.API.LogDebug("Adding user to team", "team_id", teamID, "user_id", userID)

		_, memberError := p.API.CreateTeamMember(teamID, userID)
		if memberError != nil {
			p.API.LogError("Failed to add user to team", "team_id", teamID, "user_id", userID, "error", memberError.Error())
			return false
		}
	}

	p.API.LogDebug("User is already in team", "team_id", teamID, "user_id", userID)
	return true
}

func contains(slice []string, item string) bool {
	for _, elem := range slice {
		if elem == item {
			return true
		}
	}
	return false
}
