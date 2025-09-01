package main

import (
	"encoding/json"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pkg/errors"
)

type conversationSettings struct {
	ConversationSid string `json:"conversation_sid"`
	TeamId          string `json:"team_id"`
	ChannelId       string `json:"channel_id"`
}

func (p *TwilioPlugin) getChannelConversationSid(channelId string) (string, error) {

	page := 0
	for {
		keys, err := p.API.KVList(page, 50)
		if err != nil {
			return "", errors.Wrap(err, "Could not list KV keys")
		}
		if len(keys) == 0 {
			break
		}
		for _, key := range keys {
			if strings.HasPrefix(key, "twilio-by-C-") {
				data, err := p.API.KVGet(key)
				if err != nil {
					continue
				}
				var settings conversationSettings
				if err := json.Unmarshal(data, &settings); err != nil {
					continue
				}
				if settings.ChannelId == channelId {
					return settings.ConversationSid, nil
				}
			}
		}
		page++
	}
	return "", nil
}

func (p *TwilioPlugin) createConversationSettings(conversationSid string) (*conversationSettings, error) {
	TeamId := p.getConfiguration().TeamId

	team, err := p.API.GetTeam(TeamId)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not find team with ID %s", TeamId)
	}
	var page = 0
	for channels, err := p.API.GetPublicChannelsForTeam(team.Id, page, 100); err == nil && len(channels) > 0; channels, err = p.API.GetPublicChannelsForTeam(team.Id, page, 100) {
		page++
		for _, channel := range channels {
			if channel.Props["twilio_conversation_sid"] == conversationSid {
				settings := &conversationSettings{
					ConversationSid: conversationSid,
					TeamId:          team.Id,
					ChannelId:       channel.Id,
				}
				if err := p.saveConversationSettings(settings); err != nil {
					return nil, errors.Wrap(err, "Could not save conversation settings")
				}
				return settings, nil
			}
		}
	}

	bot, appErr := p.getBot()
	if appErr != nil {
		return nil, errors.Wrap(appErr, "Could not get bot")
	}

	var channel_name string
	participants, errp := p.GetConversationParticipants(conversationSid)
	if errp != nil {
		channel_name = "Twilio Conversation " + conversationSid

	} else {
		channel_name = "Text " + strings.Join(participants, ", ")
	}

	channel := &model.Channel{
		TeamId:      team.Id,
		Type:        model.ChannelTypeOpen,
		Name:        "twilio" + strings.ToLower(conversationSid),
		DisplayName: channel_name,
		Props: map[string]interface{}{
			"twilio_conversation_sid": conversationSid,
		},
		CreatorId: bot.UserId,
	}

	channel_new, cerr := p.API.CreateChannel(channel)

	if cerr != nil {
		return nil, errors.Wrap(cerr, "Could not create channel for conversation")
	}

	settings := &conversationSettings{
		ConversationSid: conversationSid,
		TeamId:          team.Id,
		ChannelId:       channel_new.Id,
	}

	if err := p.saveConversationSettings(settings); err != nil {
		return nil, errors.Wrap(err, "Could not save conversation settings")
	}

	return settings, nil
}

func (p *TwilioPlugin) getConversationSettings(conversationSid string) (*conversationSettings, error) {
	var settings conversationSettings
	data, err := p.API.KVGet("twilio-by-C-" + conversationSid)
	if err != nil {
		return nil, errors.Wrap(err, "Could not find conversation")
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, errors.Wrap(err, "Could not unmarshal conversation settings")
	}
	return &settings, nil
}

func (p *TwilioPlugin) getOrCreateConversationSettings(conversationSid string) (*conversationSettings, error) {
	settings, err := p.getConversationSettings(conversationSid)
	if err != nil {
		return p.createConversationSettings(conversationSid)
	}
	return settings, nil
}

func (p *TwilioPlugin) saveConversationSettings(settings *conversationSettings) error {
	data, err := json.Marshal(settings)
	if err != nil {
		return errors.Wrap(err, "Could not marshal conversation settings")
	}
	if err := p.API.KVSet("twilio-by-C-"+settings.ConversationSid, data); err != nil {
		return errors.Wrap(err, "Could not save conversation settings")
	}
	return nil
}
