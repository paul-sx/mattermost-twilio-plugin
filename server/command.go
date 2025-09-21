package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	openapi "github.com/twilio/twilio-go/rest/conversations/v1"
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
		Description:      "Check to see the twilio conversation linked to this channel",
		AutoComplete:     true,
		AutoCompleteDesc: "Commands are channel, conversation, number, help",
		AutoCompleteHint: "[command]",
		AutocompleteData: getAutocompleteData(),
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

/*
	Command structure
	channel:
		status: shows conversation linked to this channel and participants
	    connect <conversation_sid>: links this channel to the given conversation
	    disconnect: unlinks this channel from any conversation
	conversation:
		list [page]: lists conversations and participants (page size 20)
		participants <conversation_sid>: lists participants in the given conversation
		webhooks:
			list <conversation_sid>: lists webhooks for the given conversation
			add <conversation_sid>: adds a webhook to the given conversation
			remove <conversation_sid>: removes the given webhook from the given conversation
	number:
		list: lists phone numbers associated with the Twilio account
		webhooks:
			setup <phone_number>: sets up a webhook for the given phone number
			remove <phone_number>: removes the webhook for the given phone number
	help: shows this message


*/

func getAutocompleteData() *model.AutocompleteData {
	main := &model.AutocompleteData{
		Trigger:  "twilio",
		Hint:     "[command]",
		HelpText: "command is one of channel, conversation, number, help",
	}
	channel := &model.AutocompleteData{
		Trigger:  "channel",
		Hint:     "[status|connect|disconnect]",
		HelpText: "channel commands are status, connect <conversation_sid>, disconnect",
	}
	channel_status := &model.AutocompleteData{
		Trigger:  "status",
		Hint:     "",
		HelpText: "shows the conversation linked to this channel",
	}
	channel.AddCommand(channel_status)
	channel_connect := &model.AutocompleteData{
		Trigger:  "connect",
		Hint:     "<conversation_sid>",
		HelpText: "links this channel to the given conversation",
	}
	channel_connect.AddTextArgument("The SID of the Twilio conversation to link to this channel", "conversation_sid", "")
	channel.AddCommand(channel_connect)
	channel_disconnect := &model.AutocompleteData{
		Trigger:  "disconnect",
		Hint:     "",
		HelpText: "unlinks this channel from any conversation",
	}
	channel.AddCommand(channel_disconnect)
	main.AddCommand(channel)

	conversation := &model.AutocompleteData{
		Trigger:  "conversation",
		Hint:     "[list|participants|webhooks]",
		HelpText: "conversation commands are list [page], participants <conversation_sid>, webhooks [list|add|remove] <conversation_sid>",
	}
	conversation_list := &model.AutocompleteData{
		Trigger:  "list",
		Hint:     "[page]",
		HelpText: "lists conversations and participants (page size 20)",
	}
	conversation_list.AddTextArgument("The page of conversations to list", "page", "")
	conversation.AddCommand(conversation_list)
	conversation_participants := &model.AutocompleteData{
		Trigger:  "participants",
		Hint:     "<conversation_sid>",
		HelpText: "lists participants in the given conversation",
	}
	conversation_participants.AddTextArgument("The SID of the Twilio conversation to list participants for", "conversation_sid", "")
	conversation.AddCommand(conversation_participants)
	conversation_webhooks := &model.AutocompleteData{
		Trigger:  "webhooks",
		Hint:     "[list|add|remove] <conversation_sid>",
		HelpText: "webhooks commands are list <conversation_sid>, add <conversation_sid>, remove <conversation_sid>",
	}
	conversation_webhooks_list := &model.AutocompleteData{
		Trigger:  "list",
		Hint:     "<conversation_sid>",
		HelpText: "lists webhooks for the given conversation",
	}
	conversation_webhooks_list.AddTextArgument("The SID of the Twilio conversation to list webhooks for", "conversation_sid", "")
	conversation_webhooks.AddCommand(conversation_webhooks_list)
	conversation_webhooks_add := &model.AutocompleteData{
		Trigger:  "add",
		Hint:     "<conversation_sid>",
		HelpText: "adds a webhook to the given conversation",
	}
	conversation_webhooks_add.AddTextArgument("The SID of the Twilio conversation to add a webhook to", "conversation_sid", "")
	conversation_webhooks.AddCommand(conversation_webhooks_add)
	conversation_webhooks_remove := &model.AutocompleteData{
		Trigger:  "remove",
		Hint:     "<conversation_sid>",
		HelpText: "removes the given webhook from the given conversation",
	}
	conversation_webhooks_remove.AddTextArgument("The SID of the Twilio conversation to remove a webhook from", "conversation_sid", "")
	conversation_webhooks.AddCommand(conversation_webhooks_remove)
	conversation.AddCommand(conversation_webhooks)
	main.AddCommand(conversation)

	number := &model.AutocompleteData{
		Trigger:  "number",
		Hint:     "[list|webhooks]",
		HelpText: "number commands are list, webhooks [setup|remove] <phone_number>",
	}
	number_list := &model.AutocompleteData{
		Trigger:  "list",
		Hint:     "",
		HelpText: "lists phone numbers associated with the Twilio account",
	}
	number.AddCommand(number_list)
	number_webhooks := &model.AutocompleteData{
		Trigger:  "webhooks",
		Hint:     "[setup|remove] <phone_number>",
		HelpText: "webhooks commands are setup <phone_number>, remove <phone_number>",
	}
	number_webhooks_setup := &model.AutocompleteData{
		Trigger:  "setup",
		Hint:     "<phone_number>",
		HelpText: "sets up a webhook for the given phone number",
	}
	number_webhooks_setup.AddTextArgument("The phone number to set up a webhook for", "phone_number", "")
	number_webhooks.AddCommand(number_webhooks_setup)
	number_webhooks_remove := &model.AutocompleteData{
		Trigger:  "remove",
		Hint:     "<phone_number>",
		HelpText: "removes the webhook for the given phone number",
	}
	number_webhooks_remove.AddTextArgument("The phone number to remove the webhook for", "phone_number", "")
	number_webhooks.AddCommand(number_webhooks_remove)
	number.AddCommand(number_webhooks)
	main.AddCommand(number)

	help := &model.AutocompleteData{
		Trigger:  "help",
		Hint:     "",
		HelpText: "shows this message",
	}
	main.AddCommand(help)

	//

	return main
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
	fields := strings.Fields(args.Command)

	if len(fields) < 2 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Available commands are channel, conversation, number, help. Use /twilio help for more information.",
		}
	}

	switch strings.ToLower(fields[1]) {
	case "channel":
		return c.executeChannelCommand(args, p, fields[2:])
	case "conversation":
		return c.executeConversationCommand(args, p, fields[2:])
	case "number":
		return c.executeNumberCommand(args, p, fields[2:])
	case "help":
		text := `**Command structure**
	**channel:**
		**status:** shows conversation linked to this channel and participants
		**connect <conversation_sid>:** links this channel to the given conversation
		**disconnect:** unlinks this channel from any conversation
	**conversation:**
		**list [page]:** lists conversations and participants (page size 20)
		**participants <conversation_sid>:** lists participants in the given conversation
		**webhooks:**
			**list <conversation_sid>:** lists webhooks for the given conversation
			**add <conversation_sid>:** adds a webhook to the given conversation
			**remove <conversation_sid>:** removes the given webhook from the given conversation
	**number:**
		**list:** lists phone numbers associated with the Twilio account
		**webhooks:**
			**setup <phone_number>:** sets up a webhook for the given phone number
			**remove <phone_number>:** removes the webhook for the given phone number
	**help:** shows this message`
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         text,
		}
	default:
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Unknown command: %s. Available commands are channel, conversation, number, help. Use /twilio help for more information.", fields[1]),
		}
	}
}

func ConversationSidIsValid(sid string) bool {
	if !strings.HasPrefix(sid, "CH") || len(sid) != 34 {
		return false
	}

	for i := 2; i < len(sid); i++ {
		c := sid[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func (c *Handler) executeChannelCommand(args *model.CommandArgs, p *TwilioPlugin, fields []string) *model.CommandResponse {
	switch strings.ToLower(fields[0]) {
	case "status":

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
	case "connect":
		if len(fields) < 2 {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "Please provide a conversation SID to connect to. Usage: /twilio channel connect <conversation_sid>",
			}
		}
		conversationSid, err := p.getChannelConversationSid(args.ChannelId)
		if err == nil || conversationSid != "" {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         fmt.Sprintf("This channel is already linked to Twilio conversation %s. Please disconnect first before connecting to a new conversation.", conversationSid),
			}
		}
		conversationSid = fields[1]
		// Test if conversationSid matches regex ^CH[0-9a-fA-F]{32}$
		if !ConversationSidIsValid(conversationSid) {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "Invalid conversation SID format. It should match ^CH[0-9a-fA-F]{32}$.",
			}
		}

		conv, err := p.twilio.GetConversation(conversationSid)
		if err != nil {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         fmt.Sprintf("Could not find Twilio conversation with SID %s.", conversationSid),
			}
		}
		settings := &conversationSettings{
			ConversationSid: conversationSid,
			TeamId:          args.TeamId,
			ChannelId:       args.ChannelId,
			ChatServiceSid:  conv.ChatServiceSid,
		}
		if err := p.saveConversationSettings(settings); err != nil {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "Could not save conversation settings.",
			}
		}
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("This channel is now linked to Twilio conversation %s.", conversationSid),
		}
	case "disconnect":
		conversation, err := p.getChannelConversationSettings(args.ChannelId)
		if err != nil || conversation == nil {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "This channel is not linked to a Twilio conversation.",
			}
		}
		p.deleteConversationSettings(conversation)
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("This channel has been unlinked from Twilio conversation %s.", conversation.ConversationSid),
		}
	}
	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "Unknown channel command. Available commands are status, connect <conversation_sid>, disconnect. Use /twilio help for more information.",
	}
}

func (c *Handler) executeConversationCommand(args *model.CommandArgs, p *TwilioPlugin, fields []string) *model.CommandResponse {
	switch strings.ToLower(fields[0]) {
	case "list":
		page := 0
		if len(fields) > 1 {
			_, err := fmt.Sscanf(fields[1], "%d", &page)
			if err != nil || page < 0 {
				return &model.CommandResponse{
					ResponseType: model.CommandResponseTypeEphemeral,
					Text:         "Invalid page number. Usage: /twilio conversation list [page]",
				}
			}
		}
		conversations, err := p.twilio.ListConversations()
		if err != nil {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "Could not list Twilio conversations.",
			}
		}
		if len(conversations) == 0 {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "No Twilio conversations found.",
			}
		}
		if page*20 >= len(conversations) {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "No more Twilio conversations found.",
			}
		}
		end := (page + 1) * 20
		if end > len(conversations) {
			end = len(conversations)
		}
		totalCount := len(conversations)
		conversations = conversations[page*20 : end]
		text := "Twilio conversations:\n"
		var wg sync.WaitGroup
		guard := make(chan struct{}, 5) // limit to 5 concurrent goroutines
		for _, conv := range conversations {
			go func(conv openapi.ConversationsV1Conversation) {
				guard <- struct{}{}
				wg.Add(1)
				defer func() {
					<-guard
					wg.Done()
				}()
				participants, err := p.twilio.GetConversationParticipants(*conv.Sid)
				if err != nil {
					text += fmt.Sprintf("- %s (could not get participants)\n", *conv.Sid)
				} else {
					text += fmt.Sprintf("- %s (participants: %s)\n", *conv.Sid, strings.Join(participants, ", "))
				}
			}(conv)
		}
		wg.Wait()
		text += fmt.Sprintf("Showing %d to %d of %d conversations. Use /twilio conversation list %d to see the next page.", page*20+1, end, totalCount, page+1)
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         text,
		}
	case "participants":
		if len(fields) < 2 {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "Please provide a conversation SID to list participants for. Usage: /twilio conversation participants <conversation_sid>",
			}
		}
		conversationSid := fields[1]
		if !ConversationSidIsValid(conversationSid) {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "Invalid conversation SID format. It should match ^CH[0-9a-fA-F]{32}$.",
			}
		}
		participants, err := p.twilio.GetConversationParticipants(conversationSid)
		if err != nil {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         fmt.Sprintf("Could not get participants for Twilio conversation %s.", conversationSid),
			}
		}
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Participants in Twilio conversation %s: %s", conversationSid, strings.Join(participants, ", ")),
		}
	case "webhooks":
		if len(fields) < 2 {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "Please provide a subcommand (list, add, remove) and a conversation SID. Usage: /twilio conversation webhooks [list|add|remove] <conversation_sid>",
			}
		}
		switch strings.ToLower(fields[1]) {
		case "list":
			if len(fields) < 3 {
				return &model.CommandResponse{
					ResponseType: model.CommandResponseTypeEphemeral,
					Text:         "Please provide a conversation SID to list webhooks for. Usage: /twilio conversation webhooks list <conversation_sid>",
				}
			}
			conversationSid := fields[2]
			if !ConversationSidIsValid(conversationSid) {
				return &model.CommandResponse{
					ResponseType: model.CommandResponseTypeEphemeral,
					Text:         "Invalid conversation SID format. It should match ^CH[0-9a-fA-F]{32}$.",
				}
			}
			webhooks, err := p.twilio.ListConversationWebhooks(conversationSid)
			if err != nil {
				return &model.CommandResponse{
					ResponseType: model.CommandResponseTypeEphemeral,
					Text:         fmt.Sprintf("Could not list webhooks for Twilio conversation %s.", conversationSid),
				}
			}
			if len(webhooks) == 0 {
				return &model.CommandResponse{
					ResponseType: model.CommandResponseTypeEphemeral,
					Text:         fmt.Sprintf("No webhooks found for Twilio conversation %s.", conversationSid),
				}
			}
			text := fmt.Sprintf("Webhooks for Twilio conversation %s:\n", conversationSid)
			for _, wh := range webhooks {
				var url string
				var eventsStr []string
				url = ""

				if wh.Configuration != nil {
					if configMap, ok := (*wh.Configuration).(map[string]interface{}); ok {
						if u, ok := configMap["url"].(string); ok {
							url = u
						}
						if events, ok := configMap["events"].([]interface{}); ok {

							for _, e := range events {
								if es, ok := e.(string); ok {
									eventsStr = append(eventsStr, es)
								}
							}
						}
					}
				}
				text += fmt.Sprintf("- SID: %s, URL: %s, Events: %s\n", *wh.Sid, url, strings.Join(eventsStr, ", "))
			}
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         text,
			}
		case "add":
			if len(fields) < 3 {
				return &model.CommandResponse{
					ResponseType: model.CommandResponseTypeEphemeral,
					Text:         "Please provide a conversation SID to add a webhook to. Usage: /twilio conversation webhooks add <conversation_sid>",
				}
			}
			conversationSid := fields[2]
			if !ConversationSidIsValid(conversationSid) {
				return &model.CommandResponse{
					ResponseType: model.CommandResponseTypeEphemeral,
					Text:         "Invalid conversation SID format. It should match ^CH[0-9a-fA-F]{32}$.",
				}
			}
			err := p.twilio.AddWebhookToConversation(conversationSid)
			if err != nil {
				return &model.CommandResponse{
					ResponseType: model.CommandResponseTypeEphemeral,
					Text:         fmt.Sprintf("Could not add webhook to Twilio conversation %s.", conversationSid),
				}
			}
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         fmt.Sprintf("Webhook added to Twilio conversation %s.", conversationSid),
			}
		case "remove":
			if len(fields) < 3 {
				return &model.CommandResponse{
					ResponseType: model.CommandResponseTypeEphemeral,
					Text:         "Please provide a conversation SID to remove the webhook from. Usage: /twilio conversation webhooks remove <conversation_sid>",
				}
			}
			conversationSid := fields[2]
			if !ConversationSidIsValid(conversationSid) {
				return &model.CommandResponse{
					ResponseType: model.CommandResponseTypeEphemeral,
					Text:         "Invalid conversation SID format. It should match ^CH[0-9a-fA-F]{32}$.",
				}
			}
			err := p.twilio.RemoveWebhookFromConversation(conversationSid)
			if err != nil {
				return &model.CommandResponse{
					ResponseType: model.CommandResponseTypeEphemeral,
					Text:         fmt.Sprintf("Could not remove webhook from Twilio conversation %s.", conversationSid),
				}
			}
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         fmt.Sprintf("Webhook removed from Twilio conversation %s.", conversationSid),
			}
		default:
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "Unknown webhooks subcommand. Available subcommands are list <conversation_sid>, add <conversation_sid>, remove <conversation_sid>. Use /twilio help for more information.",
			}
		}
	}
	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "Unknown conversation command. Available commands are list [page], participants <conversation_sid>, webhooks [list|add|remove] <conversation_sid>. Use /twilio help for more information.",
	}
}

func (c *Handler) executeNumberCommand(args *model.CommandArgs, p *TwilioPlugin, fields []string) *model.CommandResponse {

	numbers, err := p.twilio.AccountNumbers()
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Could not find phone numbers.",
		}
	}
	if len(numbers) == 0 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "No phone numbers found.",
		}
	}
	switch strings.ToLower(fields[0]) {
	case "list":
		text := "Phone numbers:\n"
		for _, num := range numbers {
			text += fmt.Sprintf("- %s\n", *num.PhoneNumber)
		}
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         text,
		}
	case "webhooks":
		if len(fields) < 2 {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "Please provide a subcommand (setup, remove) and a phone number. Usage: /twilio number webhooks [setup|remove] <phone_number>",
			}
		}
		switch strings.ToLower(fields[1]) {
		case "setup":
			if len(fields) < 3 {
				return &model.CommandResponse{
					ResponseType: model.CommandResponseTypeEphemeral,
					Text:         "Please provide a phone number to set up a webhook for. Usage: /twilio number webhooks setup <phone_number>",
				}
			}
			phoneNumber := fields[2]
			found := false
			for _, num := range numbers {
				if num.PhoneNumber != nil && *num.PhoneNumber == phoneNumber {
					found = true
					break
				}
			}
			if !found {
				return &model.CommandResponse{
					ResponseType: model.CommandResponseTypeEphemeral,
					Text:         fmt.Sprintf("Phone number %s is not associated with your Twilio account.", phoneNumber),
				}
			}
			go p.twilio.SetupPhoneNumberAsync(phoneNumber, args)
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         fmt.Sprintf("Setting up webhook for phone number %s. This may take a few seconds.", phoneNumber),
			}

		case "remove":
			if len(fields) < 3 {
				return &model.CommandResponse{
					ResponseType: model.CommandResponseTypeEphemeral,
					Text:         "Please provide a phone number to remove the webhook for. Usage: /twilio number webhooks remove <phone_number>",
				}
			}
			phoneNumber := fields[2]
			found := false
			for _, num := range numbers {
				if num.PhoneNumber != nil && *num.PhoneNumber == phoneNumber {
					found = true
					break
				}
			}
			if !found {
				return &model.CommandResponse{
					ResponseType: model.CommandResponseTypeEphemeral,
					Text:         fmt.Sprintf("Phone number %s is not associated with your Twilio account.", phoneNumber),
				}
			}
			err := p.twilio.RemovePhoneNumber(phoneNumber)
			if err != nil {
				return &model.CommandResponse{
					ResponseType: model.CommandResponseTypeEphemeral,
					Text:         fmt.Sprintf("Could not remove webhook for phone number %s.", phoneNumber),
				}
			}
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         fmt.Sprintf("Webhook removed for phone number %s.", phoneNumber),
			}
		default:
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "Unknown webhooks subcommand. Available subcommands are setup <phone_number>, remove <phone_number>. Use /twilio help for more information.",
			}
		}
	}
	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "Unknown number command. Available commands are list, webhooks [setup|remove] <phone_number>. Use /twilio help for more information.",
	}
}
