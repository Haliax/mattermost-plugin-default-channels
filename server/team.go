package main

import (
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// UserHasJoinedTeam is invoked after the membership has been committed to the database. If
// actor is not nil, the user was added to the team by the actor.
func (p *Plugin) UserHasJoinedTeam(c *plugin.Context, teamMember *model.TeamMember, actor *model.User) {
	user, err := p.API.GetUser(teamMember.UserId)
	if err != nil {
		p.API.LogError(
			"Failed to query user",
			"user_id", teamMember.UserId,
			"error", err.Error(),
		)
		return
	}

	p.addToAllDefaultChannels(user, false)
}
