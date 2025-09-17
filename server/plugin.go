package main

import (
	"net/http"
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
	commandHandler    Command
	twilio            ITwilioClient
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
	p.commandHandler = NewCommandHandler(p.client)
	bot, err := p.initializeBot()
	if err != nil {
		return err
	}
	p.bot = bot
	p.twilio = NewTwilioClient(p)
	return nil
}

func (p *TwilioPlugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	response, err := p.commandHandler.Handle(args, p)
	if err != nil {
		return nil, model.NewAppError("ExecuteCommand", "plugin.command.execute_command.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return response, nil
}

func (p *TwilioPlugin) MessageHasBeenPosted(c *plugin.Context, post *model.Post) {

	if post.IsJoinLeaveMessage() || post.IsSystemMessage() {
		return
	}

	configuration := p.getConfiguration()

	p.API.LogDebug("Message posted", "post", post)
	channel, err := p.API.GetChannel(post.ChannelId)

	if err != nil {
		return
	}
	p.API.LogDebug("Channel info", "channel", channel)
	if channel.TeamId != configuration.TeamId {
		return
	}

	sid, errs := p.getChannelConversationSid(channel.Id)
	if errs != nil || sid == "" {
		return
	}
	p.API.LogDebug("Found conversation sid", "sid", sid)
	if sentByPlugin, oks := post.GetProp("sent_by_twilio").(bool); oks && sentByPlugin {
		return
	}
	p.API.LogDebug("Sending message to conversation", "sid", sid, "message", post.Message)
	p.twilio.SendMessageToConversation(sid, post.Message)

	if len(post.FileIds) > 0 {
		for _, fileId := range post.FileIds {
			fileInfo, err := p.API.GetFileInfo(fileId)
			if err != nil {
				p.API.LogError("Failed to get file info", "fileId", fileId, "error", err.Error())
				continue
			}
			filedata, err := p.API.GetFile(fileId)
			if err != nil {
				p.API.LogError("Failed to get file data", "fileId", fileId, "error", err.Error())
				continue
			}
			p.API.LogDebug("Sending media to conversation", "sid", sid, "fileName", fileInfo.Name)
			p.twilio.SendMediaToConversation(sid, fileInfo, filedata)
		}
	}

}
