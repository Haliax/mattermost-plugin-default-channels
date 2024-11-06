package main

import "github.com/mattermost/mattermost/server/public/model"

func (p *Plugin) isDefaultChannel(channel *model.Channel, user *model.User) bool {
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
		p.addToAllDefaultChannels(user)
	}
}

func (p *Plugin) addToAllDefaultChannels(user *model.User) {
	configuration := p.getConfiguration()

	if user.IsGuest() {
		for teamID, channelIDs := range configuration.guestChannelIDs {
			for _, channelID := range channelIDs {
				channel, err := p.API.GetChannel(channelID)
				if err != nil {
					continue
				}

				if p.addUserToDefaultChannel(channel, user) {
					configuration.guestChannelIDs[teamID] = append(channelIDs, channelID)
				}
			}
		}
		return
	}

	for teamID, channelIDs := range configuration.memberChannelIDs {
		for _, channelID := range channelIDs {
			channel, err := p.API.GetChannel(channelID)
			if err != nil {
				continue
			}
			if p.addUserToDefaultChannel(channel, user) {
				configuration.memberChannelIDs[teamID] = append(channelIDs, channelID)
			}
		}
	}
}

func (p *Plugin) addUserToDefaultChannel(channel *model.Channel, user *model.User) bool {
	team, err := p.API.GetTeam(channel.TeamId)
	if err != nil {
		return false
	}

	p.ensureTeamMembership(team.Id, user.Id)
	_, err = p.API.AddUserToChannel(channel.Id, user.Id, p.botID)
	return err != nil
}

func (p *Plugin) ensureTeamMembership(teamID string, userID string) bool {
	member, teamError := p.API.GetTeamMember(teamID, userID)

	if teamError != nil || member == nil || member.DeleteAt != 0 {
		p.API.LogInfo("Adding user to team", "team_id", teamID, "user_id", userID)

		_, memberError := p.API.CreateTeamMember(teamID, userID)
		if memberError != nil {
			p.API.LogError("Failed to add user to team", "team_id", teamID, "user_id", userID, "error", memberError.Error())
			return false
		}
	}

	p.API.LogInfo("User is already in team", "team_id", teamID, "user_id", userID)
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
