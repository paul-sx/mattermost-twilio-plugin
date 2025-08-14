package main

import (
	"github.com/pkg/errors"
)

type configuration struct {
	TeamName      string
	TwilioSid     string
	TwilioToken   string
	TeamId        string
	InstallUserId string
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

	p.setConfiguration(configuration)

	return nil
}
