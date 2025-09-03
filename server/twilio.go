package main

import (
	"encoding/json"

	"github.com/twilio/twilio-go"
	twiliov1 "github.com/twilio/twilio-go/rest/conversations/v1"
)

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
