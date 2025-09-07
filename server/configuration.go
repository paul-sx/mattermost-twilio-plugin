package main

import (
	"strings"

	"github.com/pkg/errors"
)

type configuration struct {
	TeamName        string
	TwilioSid       string
	TwilioToken     string
	TeamId          string
	InstallUserId   string
	AutoAddUsers    string
	AutoAddUsersIds *[]string
	PhoneNumber     string
}

func (p *TwilioPlugin) getConfiguration() *configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &configuration{}
	}
	return p.configuration
}

func (p *TwilioPlugin) setConfiguration(configuration *configuration) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()
	p.configuration = configuration
}

func (p *TwilioPlugin) OnConfigurationChange() error {
	var configuration = new(configuration)

	if err := p.API.LoadPluginConfiguration(configuration); err != nil {
		return errors.Wrap(err, "failed to load plugin configuration")
	}
	team, err := p.API.GetTeamByName(configuration.TeamName)
	if err != nil {
		return errors.Wrapf(err, "failed to find team %s", configuration.TeamName)
	}
	configuration.TeamId = team.Id

	addUsers := strings.Split(configuration.AutoAddUsers, ",")

	configuration.AutoAddUsersIds = &[]string{}

	for userIndex, userValue := range addUsers {
		userValue = strings.TrimSpace(userValue)
		if userValue == "" {
			continue
		}
		user, uerr := p.API.GetUserByUsername(userValue)
		if uerr != nil {
			return errors.Wrapf(uerr, "failed to find user %s (index %d)", userValue, userIndex)
		}
		*configuration.AutoAddUsersIds = append(*configuration.AutoAddUsersIds, user.Id)
	}

	if configuration.TwilioSid == "" || configuration.TwilioToken == "" {
		return errors.New("Twilio SID and Token must be set")
	}

	p.setConfiguration(configuration)

	return nil
}
