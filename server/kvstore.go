package main

import (
	"encoding/json"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pkg/errors"
)

type conversationSettings struct {
	ConversationSid string  `json:"conversation_sid"`
	TeamId          string  `json:"team_id"`
	ChannelId       string  `json:"channel_id"`
	ChatServiceSid  *string `json:"chat_service_sid,omitempty"`
}

func (p *TwilioPlugin) getChannelConversationSid(channelId string) (string, error) {
	var settings conversationSettings
	data, err := p.API.KVGet("twilio-by-Ch-" + channelId)
	if err != nil {
		//Only needed until old data is all migrated.
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
						if err := p.API.KVSet("twilio-by-Co-"+settings.ConversationSid, data); err != nil {
							return settings.ConversationSid, errors.Wrap(err, "Could not save conversation settings")
						}
						if err := p.API.KVSet("twilio-by-Ch-"+settings.ChannelId, data); err != nil {
							return settings.ConversationSid, errors.Wrap(err, "Could not save conversation settings by channel")
						}
						if err := p.API.KVDelete(key); err != nil {
							return settings.ConversationSid, errors.Wrap(err, "Could not delete old conversation settings")
						}
						return settings.ConversationSid, nil
					}
				}
			}
			page++
		}
		return "", errors.Wrap(err, "Could not find conversation")
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		return "", errors.Wrap(err, "Could not unmarshal conversation settings")
	}
	return settings.ConversationSid, nil

}

func (p *TwilioPlugin) createConversationSettings(conversationSid string) (*conversationSettings, error) {
	configuration := p.getConfiguration()

	TeamId := configuration.TeamId

	team, err := p.API.GetTeam(TeamId)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not find team with ID %s", TeamId)
	}

	bot, appErr := p.getBot()
	if appErr != nil {
		return nil, errors.Wrap(appErr, "Could not get bot")
	}

	var channel_name string
	participants, errp := p.twilio.GetConversationParticipants(conversationSid)
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

	if configuration.AutoAddUsersIds != nil {
		for _, userId := range *configuration.AutoAddUsersIds {
			if _, err := p.API.AddUserToChannel(channel_new.Id, userId, userId); err != nil {
				p.API.LogError("Could not add user to channel", "user_id", userId, "channel_id", channel_new.Id, "error", err.Error())
			}
		}
	}

	conv, errc := p.twilio.GetConversation(conversationSid)
	if errc != nil {
		return nil, errors.Wrap(errc, "Could not get conversation details")
	}

	var chatServiceSid *string
	if conv.ChatServiceSid != nil {
		chatServiceSid = conv.ChatServiceSid
	}

	settings := &conversationSettings{
		ConversationSid: conversationSid,
		TeamId:          team.Id,
		ChannelId:       channel_new.Id,
		ChatServiceSid:  chatServiceSid,
	}

	if err := p.saveConversationSettings(settings); err != nil {
		return nil, errors.Wrap(err, "Could not save conversation settings")
	}

	return settings, nil
}

func (p *TwilioPlugin) getConversationSettings(conversationSid string) (*conversationSettings, error) {
	var settings conversationSettings
	data, err := p.API.KVGet("twilio-by-Co-" + conversationSid)
	if err != nil {
		return nil, errors.Wrap(err, "Could not find conversation")
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, errors.Wrap(err, "Could not unmarshal conversation settings")
	}
	if settings.ChatServiceSid == nil {
		conv, errc := p.twilio.GetConversation(conversationSid)
		if errc != nil {
			return nil, errors.Wrap(errc, "Could not get conversation details")
		}
		if conv.ChatServiceSid != nil {
			settings.ChatServiceSid = conv.ChatServiceSid
			if err := p.saveConversationSettings(&settings); err != nil {
				return nil, errors.Wrap(err, "Could not save updated conversation settings")
			}
		}
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
	if err := p.API.KVSet("twilio-by-Co-"+settings.ConversationSid, data); err != nil {
		return errors.Wrap(err, "Could not save conversation settings")
	}
	if err := p.API.KVSet("twilio-by-Ch-"+settings.ChannelId, data); err != nil {
		return errors.Wrap(err, "Could not save conversation settings by channel")
	}
	return nil
}
