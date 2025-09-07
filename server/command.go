package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/pluginapi"
)

type Handler struct {
	client *pluginapi.Client
}

type Command interface {
	Handle(args *model.CommandArgs, p *TwilioPlugin) (*model.CommandResponse, error)
	executeTwilioCommand(args *model.CommandArgs, p *TwilioPlugin) *model.CommandResponse
}

func NewCommandHandler(client *pluginapi.Client) Command {
	err := client.SlashCommand.Register(&model.Command{
		Trigger:          "twilio",
		DisplayName:      "Twilio",
		Description:      "Turn ntfy notifications on or off for a channel or set the topic",
		AutoComplete:     true,
		AutoCompleteDesc: "Turn ntfy notifications on or off for a channel or set the topic",
		AutoCompleteHint: "on|off|topic [topic]|delay [seconds]",
		IconURL:          "https://ntfy.sh/static/images/favicon.ico",
	})
	if err != nil {
		client.Log.Error("Failed to register slash command", "error", err)
	}
	// Return command handler
	return &Handler{
		client: client,
	}
}

func (c *Handler) Handle(args *model.CommandArgs, p *TwilioPlugin) (*model.CommandResponse, error) {
	fields := strings.Fields(args.Command)
	trigger := strings.TrimPrefix(fields[0], "/")
	//if trigger != ntfyCommandTrigger {

	if trigger != "twilio" {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Unknown command: %s", args.Command),
		}, nil
	}
	return c.executeTwilioCommand(args, p), nil
}

func (c *Handler) executeTwilioCommand(args *model.CommandArgs, p *TwilioPlugin) *model.CommandResponse {

	conversationSid, err := p.getChannelConversationSid(args.ChannelId)

	if err != nil || conversationSid == "" {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "This channel is not linked to a Twilio conversation.",
		}
	}
	participants, err := p.twilio.GetConversationParticipants(conversationSid)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Could not get Twilio conversation participants.",
		}
	}
	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         fmt.Sprintf("This channel is linked to Twilio conversation %s with participants: %s", conversationSid, strings.Join(participants, ", ")),
	}

}
