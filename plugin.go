package main

import (
	"regexp"
	"sync"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
)

type Plugin struct {
	plugin.MattermostPlugin

	client            *pluginapi.Client
	configurationLock sync.RWMutex
	configuration     *configuration
}

func (p *Plugin) OnActivate() error {
	p.client = pluginapi.NewClient(p.API, p.Driver)

	return nil
}

func (p *Plugin) MessageHasBeenPosted(c *plugin.Context, post *model.Post) {
	configuration := p.getConfiguration()

	channel, err := p.API.GetChannel(post.ChannelId)

	if err != nil {
		return
	}

	if channel.TeamId != configuration.TeamId {
		return
	}

	if sentByPlugin, _ := post.GetProp("sent_by_twilio").(bool); sentByPlugin {
		return
	}

	// Limited here to Country Code +1
	pattern := `<([+]1\d{10})>`
	re := regexp.MustCompile(pattern)

	match := re.FindStringSubmatch(channel.Name)

	if match == nil {
		return
	}

}
