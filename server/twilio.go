package main

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/twilio/twilio-go"
	twiliov1 "github.com/twilio/twilio-go/rest/conversations/v1"
)

type ITwilioClient interface {
	GetConversationParticipants(conversationSid string) ([]string, error)
	SendMessageToConversation(conversationSid, message string) error
	AddWebhookToConversation(conversationSid string) error
	RemoveWebhookFromConversation(conversationSid string) error
	SetupPhoneNumber(phoneNumber string) error
	GetConversationServices() ([]twiliov1.ConversationsV1Service, error)
	CheckServiceWebhook(serviceSid string) (bool, error)
}

type TwilioClient struct {
	p       *TwilioPlugin
	client  *twilio.RestClient
	webhook string
}

func NewTwilioClient(p *TwilioPlugin) ITwilioClient {
	config := p.getConfiguration()

	clientParams := twilio.ClientParams{Username: config.TwilioSid, Password: config.TwilioToken}
	client := twilio.NewRestClientWithParams(clientParams)

	webhook, _ := url.JoinPath(*p.API.GetConfig().ServiceSettings.SiteURL, "/plugins/sx.paul.mattermost.twilio/twilio/conversation")
	return &TwilioClient{
		p:       p,
		client:  client,
		webhook: webhook,
	}
}

func (tc *TwilioClient) GetConversationParticipants(conversationSid string) ([]string, error) {

	tc.p.API.LogDebug("Getting participants for conversation", "sid", conversationSid)

	var participants []string
	params := &twiliov1.ListConversationParticipantParams{}
	resp, err := tc.client.ConversationsV1.ListConversationParticipant(conversationSid, params)
	if err != nil {
		tc.p.API.LogError("Error getting participants for conversation", "sid", conversationSid, "error", err.Error())
		return nil, err
	}
	for _, participant := range resp {
		jp, jperr := json.Marshal(participant)
		jpStr := string(jp)
		if jperr != nil {
			tc.p.API.LogError("Error marshalling participant", "participant", participant, "error", jperr.Error())
		}
		tc.p.API.LogDebug("Found participant", "participant", participant, "json", jpStr)
		tc.p.API.LogDebug("Participant binding", "binding", *participant.MessagingBinding)
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
				if addr, ok := mbMap["projected_address"].(string); ok && addr != "" {
					participants = append(participants, "*"+addr)
				}
				if addr, ok := mbMap["author_address"].(string); ok && addr != "" {
					participants = append(participants, addr)
				}
			}
		}
	}
	tc.p.API.LogDebug("Got participants for conversation", "sid", conversationSid, "participants", participants)
	return participants, nil
}

func (tc *TwilioClient) SendMessageToConversation(conversationSid, message string) error {
	tc.p.API.LogDebug("Sending message to conversation", "sid", conversationSid, "message", message)

	params := &twiliov1.CreateConversationMessageParams{Body: &message}
	_, err := tc.client.ConversationsV1.CreateConversationMessage(conversationSid, params)
	if err != nil {
		tc.p.API.LogError("Error sending message to conversation", "sid", conversationSid, "message", message, "error", err.Error())
	}
	return err
}

func (tc *TwilioClient) GetConversationServices() ([]twiliov1.ConversationsV1Service, error) {
	var services []twiliov1.ConversationsV1Service
	//resp, err := tc.client.ConversationsV1.StreamService()
	params := &twiliov1.ListServiceParams{}
	resp, err := tc.client.ConversationsV1.ListService(params)
	if err != nil {
		tc.p.API.LogError("Error getting conversation services", "error", err.Error())
		return nil, err
	}
	services = append(services, resp...)

	return services, nil
}

func (tc *TwilioClient) CheckServiceWebhook(serviceSid string) (bool, error) {

	resp, err := tc.client.ConversationsV1.FetchServiceWebhookConfiguration(serviceSid)
	if err != nil {
		tc.p.API.LogError("Error getting service webhooks", "service_sid", serviceSid, "error", err.Error())
		return false, err
	}
	return resp != nil && *resp.PostWebhookUrl == tc.webhook, nil
}

/*
 Setup phone number to use the plugin webhook
 1. Searches for existing conversations with the phone number
 2. Adds the plugin webhook for each conversation found
 3. Updates the phone number to auto create conversations
 4. Sets the plugin webhook to receive new conversations
*/

func (tc *TwilioClient) AddWebhookToConversation(conversationSid string) error {

	resp, err := tc.client.ConversationsV1.ListConversationScopedWebhook(conversationSid, &twiliov1.ListConversationScopedWebhookParams{})
	if err != nil {
		tc.p.API.LogError("Error getting conversation webhooks", "conversation_sid", conversationSid, "error", err.Error())
		return err
	}
	for _, webhook := range resp {
		if webhook.Url != nil && strings.EqualFold(*webhook.Url, tc.webhook) {
			tc.p.API.LogDebug("Webhook already exists for conversation", "conversation_sid", conversationSid, "webhook_sid", *webhook.Sid)
			return nil
		}
	}

	params := &twiliov1.CreateConversationScopedWebhookParams{}
	params.SetConfigurationMethod("POST")
	params.SetConfigurationUrl(tc.webhook)
	params.SetConfigurationFilters([]string{"onMessageAdded"})
	params.SetTarget("webhook")

	_, err = tc.client.ConversationsV1.CreateConversationScopedWebhook(conversationSid, params)
	if err != nil {
		tc.p.API.LogError("Error creating conversation webhook", "conversation_sid", conversationSid, "error", err.Error())
		return err
	}
	return nil
	//
}

func (tc *TwilioClient) RemoveWebhookFromConversation(conversationSid string) error {

	resp, err := tc.client.ConversationsV1.ListConversationScopedWebhook(conversationSid, &twiliov1.ListConversationScopedWebhookParams{})
	if err != nil {
		tc.p.API.LogError("Error getting conversation webhooks", "conversation_sid", conversationSid, "error", err.Error())
		return err
	}
	for _, webhook := range resp {
		if webhook.Url != nil && strings.EqualFold(*webhook.Url, tc.webhook) {
			// Delete the webhook
			err = tc.client.ConversationsV1.DeleteConversationScopedWebhook(conversationSid, *webhook.Sid)
			if err != nil {
				tc.p.API.LogError("Error deleting conversation webhook", "conversation_sid", conversationSid, "webhook_sid", *webhook.Sid, "error", err.Error())
				return err
			}
			tc.p.API.LogDebug("Deleted webhook from conversation", "conversation_sid", conversationSid, "webhook_sid", *webhook.Sid)
		}
	}
	return nil
}

func (tc *TwilioClient) FindConversationsByProxyAddress(proxyAddress string) ([]twiliov1.ConversationsV1Conversation, error) {

	var conversations []twiliov1.ConversationsV1Conversation
	params := &twiliov1.ListConversationParams{}
	resp, err := tc.client.ConversationsV1.ListConversation(params)
	if err != nil {
		tc.p.API.LogError("Error getting conversations", "error", err.Error())
		return nil, err
	}
	for _, conversation := range resp {
		// Check if the conversation has a participant with the proxy address
		participants, err := tc.GetConversationParticipants(*conversation.Sid)
		if err != nil {
			tc.p.API.LogError("Error getting participants for conversation", "sid", *conversation.Sid, "error", err.Error())
			continue
		}
		for _, participant := range participants {
			if strings.EqualFold(participant, "*"+proxyAddress) {
				conversations = append(conversations, conversation)
				break
			}
		}
	}

	return conversations, nil
}

func (tc *TwilioClient) SetupPhoneNumber(phoneNumber string) error {

	// Find conversations with the phone number as a participant
	conversations, err := tc.FindConversationsByProxyAddress(phoneNumber)
	if err != nil {
		return err
	}
	for _, conversation := range conversations {
		// Add webhook to conversation
		err := tc.AddWebhookToConversation(*conversation.Sid)
		if err != nil {
			return err
		}
	}

	// Update the phone number to auto create conversations and set the webhook for new conversations
	resp, err := tc.client.ConversationsV1.FetchConfigurationAddress(phoneNumber)

	if err != nil || resp == nil || resp.Sid == nil {
		params := &twiliov1.CreateConfigurationAddressParams{}
		params.SetType("sms")
		params.SetAddress(phoneNumber)
		params.SetAutoCreationEnabled(true)
		params.SetAutoCreationType("webhook")
		params.SetAutoCreationWebhookMethod("POST")
		params.SetAutoCreationWebhookUrl(tc.webhook)
		params.SetAutoCreationWebhookFilters([]string{"onConversationAdded", "onMessageAdded"})
		respc, errc := tc.client.ConversationsV1.CreateConfigurationAddress(params)
		if errc != nil {
			tc.p.API.LogError("Error creating phone number configuration", "phone_number", phoneNumber, "error", errc.Error())
			return errc
		}
		tc.p.API.LogDebug("Created phone number configuration", "phone_number", phoneNumber, "sid", *respc.Sid)
		return nil
	}

	params := &twiliov1.UpdateConfigurationAddressParams{}
	params.SetAutoCreationEnabled(true)
	params.SetAutoCreationType("webhook")
	params.SetAutoCreationWebhookMethod("POST")
	params.SetAutoCreationWebhookUrl(tc.webhook)
	params.SetAutoCreationWebhookFilters([]string{"onConversationAdded", "onMessageAdded"})
	_, err = tc.client.ConversationsV1.UpdateConfigurationAddress(*resp.Sid, params)
	if err != nil {
		tc.p.API.LogError("Error updating phone number configuration", "phone_number", phoneNumber, "error", err.Error())
		return err
	}
	tc.p.API.LogDebug("Updated phone number configuration", "phone_number", phoneNumber)
	return nil
}
