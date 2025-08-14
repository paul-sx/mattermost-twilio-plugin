package main

import (
	"sync"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
)

type TwilioPlugin struct {
	plugin.MattermostPlugin

	router *mux.Router

	client            *pluginapi.Client
	configurationLock sync.RWMutex
	configuration     *configuration
	bot               *twilioBot
}

func (p *TwilioPlugin) OnInstall(c *plugin.Context, event model.OnInstallEvent) error {
	config := p.getConfiguration()
	config.InstallUserId = event.UserId
	p.setConfiguration(config)
	return nil
}

func (p *TwilioPlugin) OnActivate() error {
	p.client = pluginapi.NewClient(p.API, p.Driver)
	p.initializeRouter()
	bot, err := p.initializeBot()
	if err != nil {
		return err
	}
	p.bot = bot
	return nil
}

func (p *TwilioPlugin) MessageHasBeenPosted(c *plugin.Context, post *model.Post) {
	configuration := p.getConfiguration()

	channel, err := p.API.GetChannel(post.ChannelId)

	if err != nil {
		return
	}

	if channel.TeamId != configuration.TeamId {
		return
	}

	if conversationSid, ok := channel.Props["twilio_conversation_sid"]; !ok || conversationSid == "" {
		return
	}

	if sentByPlugin, _ := post.GetProp("sent_by_twilio").(bool); sentByPlugin {
		return
	}

}
