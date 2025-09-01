package main

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/twilio/twilio-go"
	twiliov1 "github.com/twilio/twilio-go/rest/conversations/v1"
)

type TwilioPlugin struct {
	plugin.MattermostPlugin

	router *mux.Router

	client            *pluginapi.Client
	configurationLock sync.RWMutex
	configuration     *configuration
	bot               *twilioBot
	commandHandler    Command
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
	p.SendMessageToConversation(sid, post.Message)

}

func (p *TwilioPlugin) GetConversationParticipants(conversationSid string) ([]string, error) {

	p.API.LogDebug("Getting participants for conversation", "sid", conversationSid)

	config := p.getConfiguration()

	clientParams := twilio.ClientParams{Username: config.TwilioSid, Password: config.TwilioToken}
	client := twilio.NewRestClientWithParams(clientParams)

	var participants []string
	params := &twiliov1.ListConversationParticipantParams{}
	resp, err := client.ConversationsV1.ListConversationParticipant(conversationSid, params)
	if err != nil {
		p.API.LogError("Error getting participants for conversation", "sid", conversationSid, "error", err.Error())
		return nil, err
	}
	for _, participant := range resp {
		jp, jperr := json.Marshal(participant)
		jpStr := string(jp)
		if jperr != nil {
			p.API.LogError("Error marshalling participant", "participant", participant, "error", jperr.Error())
		}
		p.API.LogDebug("Found participant", "participant", participant, "json", jpStr)
		p.API.LogDebug("Participant binding", "binding", *participant.MessagingBinding)
		binding := participant.MessagingBinding
		if binding != nil {
			// MessagingBinding is a map[string]interface{}, try to get "address"
			if mbMap, ok := (*binding).(map[string]interface{}); ok {
				if addr, ok := mbMap["address"].(string); ok && addr != "" {
					participants = append(participants, addr)
				}
				if addr, ok := mbMap["proxy_address"].(string); ok && addr != "" {
					participants = append(participants, "*"+addr)
				}
			}
		}
	}
	p.API.LogDebug("Got participants for conversation", "sid", conversationSid, "participants", participants)
	return participants, nil
}

func (p *TwilioPlugin) SendMessageToConversation(conversationSid, message string) error {
	p.API.LogDebug("Sending message to conversation", "sid", conversationSid, "message", message)
	config := p.getConfiguration()

	clientParams := twilio.ClientParams{Username: config.TwilioSid, Password: config.TwilioToken}
	client := twilio.NewRestClientWithParams(clientParams)
	params := &twiliov1.CreateConversationMessageParams{Body: &message}
	_, err := client.ConversationsV1.CreateConversationMessage(conversationSid, params)
	if err != nil {
		p.API.LogError("Error sending message to conversation", "sid", conversationSid, "message", message, "error", err.Error())
	}
	return err
}
