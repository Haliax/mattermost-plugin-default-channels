package main

import (
	"reflect"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"

	"github.com/pkg/errors"
)

// configuration captures the plugin's external configuration as exposed in the Mattermost server
// configuration, as well as values computed from the configuration. Any public fields will be
// deserialized from the Mattermost server configuration in OnConfigurationChange.
//
// As plugins are inherently concurrent (hooks being called asynchronously), and the plugin
// configuration can change at any time, access to the configuration must be synchronized. The
// strategy used in this plugin is to guard a pointer to the configuration, and clone the entire
// struct whenever it changes. You may replace this with whatever strategy you choose.
//
// If you add non-reference types to your configuration struct, be sure to rewrite Clone as a deep
// copy appropriate for your types.
type configuration struct {
	// A list of channels to which users are automatically added.
	MemberChannelNames string

	// A list of channels to which guests are automatically added.
	GuestChannelNames string

	// If false, ignore teams the users is not a member of.
	AddToTeam bool

	// demoUserID is the id of the user specified above.
	demoUserID string

	// memberChannelIDs maps team ids to the channel ids.
	memberChannelIDs map[string][]string

	// memberChannelIDs maps team ids to the channel ids.
	guestChannelIDs map[string][]string
}

// Clone shallow copies the configuration. Your implementation may require a deep copy if
// your configuration has reference types.
func (c *configuration) Clone() *configuration {
	// Deep copy memberChannelIDs and guestChannelIDs, a reference type.
	memberChannelIDs := make(map[string][]string)
	for key, value := range c.memberChannelIDs {
		memberChannelIDs[key] = value
	}

	guestChannelIDs := make(map[string][]string)
	for key, value := range c.guestChannelIDs {
		guestChannelIDs[key] = value
	}

	return &configuration{
		MemberChannelNames: c.MemberChannelNames,
		GuestChannelNames:  c.GuestChannelNames,
		AddToTeam:          c.AddToTeam,
		memberChannelIDs:   memberChannelIDs,
		guestChannelIDs:    guestChannelIDs,
	}
}

// getConfiguration retrieves the active configuration under lock, making it safe to use
// concurrently. The active configuration may change underneath the client of this method, but
// the struct returned by this API call is considered immutable.
func (p *Plugin) getConfiguration() *configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &configuration{}
	}

	return p.configuration
}

// setConfiguration replaces the active configuration under lock.
//
// Do not call setConfiguration while holding the configurationLock, as sync.Mutex is not
// reentrant. In particular, avoid using the plugin API entirely, as this may in turn trigger a
// hook back into the plugin. If that hook attempts to acquire this lock, a deadlock may occur.
//
// This method panics if setConfiguration is called with the existing configuration. This almost
// certainly means that the configuration was modified without being cloned and may result in
// an unsafe access.
func (p *Plugin) setConfiguration(configuration *configuration) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()

	if configuration != nil && p.configuration == configuration {
		// Ignore assignment if the configuration struct is empty. Go will optimize the
		// allocation for same to point at the same memory address, breaking the check
		// above.
		if reflect.ValueOf(*configuration).NumField() == 0 {
			return
		}

		panic("setConfiguration called with the existing configuration")
	}

	p.configuration = configuration
}

func (p *Plugin) generateChannelList(channelString string) map[string][]string {
	channelList := make(map[string][]string)

	if channelString == "" {
		return channelList
	}

	splicedList := strings.Split(channelString, ",")

	for _, memberChannel := range splicedList {
		channelDetails := strings.Split(memberChannel, ":")
		if len(channelDetails) != 2 {
			p.API.LogWarn("default channel plugin - could not identify channel: " + memberChannel)
			continue
		}

		teamName := channelDetails[0]
		channelName := channelDetails[1]

		channel, _ := p.API.GetChannelByNameForTeamName(teamName, channelName, false)
		if channel == nil {
			p.API.LogWarn("default channel plugin - could not find channel: " + memberChannel)
			continue
		}

		channelList[channel.TeamId] = append(channelList[channel.TeamId], channel.Id)
		p.API.LogInfo("default channel plugin - added default channel: " + teamName + ":" + channelName)
	}

	return channelList
}

// OnConfigurationChange is invoked when configuration changes may have been made.
func (p *Plugin) OnConfigurationChange() error {
	configuration := p.getConfiguration().Clone()

	// Load the public configuration fields from the Mattermost server configuration.
	if err := p.API.LoadPluginConfiguration(configuration); err != nil {
		return errors.Wrap(err, "failed to load plugin configuration")
	}

	botID, ensureBotError := p.API.EnsureBotUser(&model.Bot{
		Username:    "default-channel-bot",
		DisplayName: "Default Channel Bot",
		Description: "A bot ensuring you are in all default channels.",
	})
	if ensureBotError != nil {
		return errors.Wrap(ensureBotError, "failed to ensure default channel bot")
	}

	p.botID = botID

	// Extract list of channels for each team.
	configuration.memberChannelIDs = p.generateChannelList(configuration.MemberChannelNames)
	configuration.guestChannelIDs = p.generateChannelList(configuration.GuestChannelNames)

	p.setConfiguration(configuration)

	p.addAllUsersToDefaultChannels()

	return nil
}
