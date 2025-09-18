package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pkg/errors"
	"github.com/twilio/twilio-go"
	twiliov1 "github.com/twilio/twilio-go/rest/conversations/v1"
	messaging "github.com/twilio/twilio-go/rest/messaging/v1"
)

type ITwilioClient interface {
	GetConversationParticipants(conversationSid string) ([]string, error)
	GetConversation(conversationSid string) (*twiliov1.ConversationsV1Conversation, error)
	SendMessageToConversation(conversationSid, message string) error
	SendMediaToConversation(conversationSid string, media *model.FileInfo, mediadata []byte) error
	ListConversationWebhooks(conversationSid string) ([]twiliov1.ConversationsV1ConversationScopedWebhook, error)
	AddWebhookToConversation(conversationSid string) error
	RemoveWebhookFromConversation(conversationSid string) error
	SetupPhoneNumber(phoneNumber string) error
	RemovePhoneNumber(phoneNumber string) error
	AccountNumbers() ([]messaging.MessagingV1PhoneNumber, error)
	AccountNumbersStrings() ([]string, error)
	GetConversationServices() ([]twiliov1.ConversationsV1Service, error)
	CheckServiceWebhook(serviceSid string) (bool, error)
	FindConversationsByProxyAddress(proxyAddress string) ([]twiliov1.ConversationsV1Conversation, error)
	DownloadMedia(ChatServiceSid string, mediaSid string) ([]byte, error)
	ListConversations() ([]twiliov1.ConversationsV1Conversation, error)
}

type TwilioClient struct {
	p       *TwilioPlugin
	client  *twilio.RestClient
	webhook string
}

func (tc *TwilioClient) DownloadMedia(ChatServiceSid string, mediaSid string) ([]byte, error) {
	req, err := http.NewRequest("GET", "https://mcs.us1.twilio.com/v1/Services/"+ChatServiceSid+"/Media/"+mediaSid+"/Content", nil)
	if err != nil {
		return nil, err
	}
	config := tc.p.getConfiguration()
	req.SetBasicAuth(config.TwilioSid, config.TwilioToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New("failed to download media, status code: " + resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
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

func (tc *TwilioClient) GetConversation(conversationSid string) (*twiliov1.ConversationsV1Conversation, error) {

	tc.p.API.LogDebug("Getting conversation", "sid", conversationSid)

	resp, err := tc.client.ConversationsV1.FetchConversation(conversationSid)
	if err != nil {
		tc.p.API.LogError("Error getting conversation", "sid", conversationSid, "error", err.Error())
		return nil, err
	}
	return resp, nil
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

func (tc *TwilioClient) SendMessageToConversation(conversationSid string, message string) error {
	tc.p.API.LogDebug("Sending message to conversation", "sid", conversationSid, "message", message)

	params := &twiliov1.CreateConversationMessageParams{Body: &message}
	_, err := tc.client.ConversationsV1.CreateConversationMessage(conversationSid, params)
	if err != nil {
		tc.p.API.LogError("Error sending message to conversation", "sid", conversationSid, "message", message, "error", err.Error())
	}
	return err
}

func (tc *TwilioClient) SendMediaToConversation(conversationSid string, media *model.FileInfo, mediadata []byte) error {
	settings, err := tc.p.getConversationSettings(conversationSid)
	if err != nil {
		tc.p.API.LogError("Could not get conversation settings", "sid", conversationSid, "error", err.Error())
		return err
	}
	if settings.ChatServiceSid == nil {
		tc.p.API.LogError("Conversation does not have a chat service sid", "sid", conversationSid)
		return errors.New("conversation does not have a chat service sid")
	}

	tc.p.API.LogDebug("Sending media to conversation", "sid", conversationSid, "media", media.Name)

	// Upload the media to Twilio Media Content Service
	req, err := http.NewRequest("POST", "https://mcs.us1.twilio.com/v1/Services/"+*settings.ChatServiceSid+"/Media", strings.NewReader(string(mediadata)))
	if err != nil {
		tc.p.API.LogError("Error creating request to upload media", "error", err.Error())
		return err
	}
	config := tc.p.getConfiguration()
	req.SetBasicAuth(config.TwilioSid, config.TwilioToken)
	req.Header.Set("Content-Type", media.MimeType)
	req.Header.Set("Content-Length", strconv.Itoa(len(mediadata)))
	req.Header.Set("X-Twilio-File-Name", media.Name)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		tc.p.API.LogError("Error uploading media to Twilio", "error", err.Error())
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		tc.p.API.LogError("Error uploading media to Twilio, non-2xx response", "status", resp.StatusCode, "body", string(body))
		return errors.New("failed to upload media to Twilio, status code: " + resp.Status)
	}
	var uploadResp struct {
		Sid string `json:"sid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		tc.p.API.LogError("Error decoding upload media response", "error", err.Error())
		return err
	}
	if uploadResp.Sid == "" {
		tc.p.API.LogError("Upload media response did not contain a sid")
		return errors.New("upload media response did not contain a sid")
	}

	// Send the media message to the conversation
	//mediaUrl := "https://mcs.us1.twilio.com/v1/Services/" + *settings.ChatServiceSid + "/Media/" + uploadResp.Sid
	params := &twiliov1.CreateConversationMessageParams{}
	params.SetMediaSid(uploadResp.Sid)

	_, err = tc.client.ConversationsV1.CreateConversationMessage(conversationSid, params)
	if err != nil {
		tc.p.API.LogError("Error sending media message to conversation", "sid", conversationSid, "media_sid", uploadResp.Sid, "error", err.Error())
	}
	return nil
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

func (tc *TwilioClient) ListConversationWebhooks(conversationSid string) ([]twiliov1.ConversationsV1ConversationScopedWebhook, error) {

	var webhooks []twiliov1.ConversationsV1ConversationScopedWebhook
	params := &twiliov1.ListConversationScopedWebhookParams{}
	resp, err := tc.client.ConversationsV1.ListConversationScopedWebhook(conversationSid, params)
	if err != nil {
		tc.p.API.LogError("Error getting conversation webhooks", "conversation_sid", conversationSid, "error", err.Error())
		return nil, err
	}
	webhooks = append(webhooks, resp...)
	return webhooks, nil
}

func (tc *TwilioClient) AddWebhookToConversation(conversationSid string) error {

	resp, err := tc.client.ConversationsV1.ListConversationScopedWebhook(conversationSid, &twiliov1.ListConversationScopedWebhookParams{})
	if err != nil {
		tc.p.API.LogError("Error getting conversation webhooks", "conversation_sid", conversationSid, "error", err.Error())
		return err
	}
	for _, webhook := range resp {
		var url string
		url = ""
		if webhook.Configuration != nil {
			if configMap, ok := (*webhook.Configuration).(map[string]interface{}); ok {
				if u, ok := configMap["url"].(string); ok {
					url = u
				}
			}
		}
		if strings.EqualFold(url, tc.webhook) {
			tc.p.API.LogDebug("Webhook already exists for conversation", "conversation_sid", conversationSid, "webhook_sid", *webhook.Sid)
			return nil
		}
	}

	params := &twiliov1.CreateConversationScopedWebhookParams{}
	params.SetConfigurationMethod("post")
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
		var url string
		url = ""
		if webhook.Configuration != nil {
			if configMap, ok := (*webhook.Configuration).(map[string]interface{}); ok {
				if u, ok := configMap["url"].(string); ok {
					url = u
				}
			}
		}

		if strings.EqualFold(url, tc.webhook) {
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

func (tc *TwilioClient) ListConversations() ([]twiliov1.ConversationsV1Conversation, error) {
	var conversations []twiliov1.ConversationsV1Conversation
	params := &twiliov1.ListConversationParams{}
	resp, err := tc.client.ConversationsV1.ListConversation(params)
	if err != nil {
		tc.p.API.LogError("Error getting conversations", "error", err.Error())
		return nil, err
	}
	conversations = append(conversations, resp...)
	return conversations, nil
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

	// Update the phone number to auto create conversations and set the webhook for new conversations
	resp, err := tc.client.ConversationsV1.FetchConfigurationAddress(phoneNumber)

	if err != nil || resp == nil || resp.Sid == nil {
		params := &twiliov1.CreateConfigurationAddressParams{}
		params.SetType("sms")
		params.SetAddress(phoneNumber)
		params.SetAutoCreationEnabled(true)
		params.SetAutoCreationType("webhook")
		params.SetAutoCreationWebhookMethod("post")
		params.SetAutoCreationWebhookUrl(tc.webhook)
		params.SetAutoCreationWebhookFilters([]string{"onMessageAdded"})
		respc, errc := tc.client.ConversationsV1.CreateConfigurationAddress(params)
		if errc != nil {
			tc.p.API.LogError("Error creating phone number configuration", "phone_number", phoneNumber, "error", errc.Error())
			return errc
		}
		tc.p.API.LogDebug("Created phone number configuration", "phone_number", phoneNumber, "sid", *respc.Sid)

	} else {
		params := &twiliov1.UpdateConfigurationAddressParams{}
		params.SetAutoCreationEnabled(true)
		params.SetAutoCreationType("webhook")
		params.SetAutoCreationWebhookMethod("post")
		params.SetAutoCreationWebhookUrl(tc.webhook)
		params.SetAutoCreationWebhookFilters([]string{"onMessageAdded"})
		_, err = tc.client.ConversationsV1.UpdateConfigurationAddress(*resp.Sid, params)
		if err != nil {
			tc.p.API.LogError("Error updating phone number configuration", "phone_number", phoneNumber, "error", err.Error())
			return err
		}
		tc.p.API.LogDebug("Updated phone number configuration", "phone_number", phoneNumber)
	}

	// Find conversations with the phone number as a participant
	tc.p.API.LogDebug("Setting up phone number:", phoneNumber)
	conversations, err := tc.FindConversationsByProxyAddress(phoneNumber)
	if err != nil {
		return err
	}
	for _, conversation := range conversations {
		// Add webhook to conversation
		tc.p.API.LogDebug("Adding webhook to conversation:", "conversation", *conversation.Sid)
		err := tc.AddWebhookToConversation(*conversation.Sid)
		if err != nil {
			tc.p.API.LogError("Error adding webhook to conversation:", *conversation.Sid, "error:", err.Error())
			return err
		}
	}
	return nil

}

func (tc *TwilioClient) RemovePhoneNumber(phoneNumber string) error {

	// Find conversations with the phone number as a participant
	tc.p.API.LogDebug("Removing phone number:", "number", phoneNumber)
	conversations, err := tc.FindConversationsByProxyAddress(phoneNumber)
	if err != nil {
		return err
	}
	for _, conversation := range conversations {
		// Add webhook to conversation
		tc.p.API.LogDebug("Removing webhook to conversation:", "sid", *conversation.Sid)
		err := tc.RemoveWebhookFromConversation(*conversation.Sid)
		if err != nil {
			return err
		}
	}
	tc.p.API.LogDebug("Fetching configuration for phone number:", "number", phoneNumber)
	// Update the phone number to auto create conversations and set the webhook for new conversations
	resp, err := tc.client.ConversationsV1.FetchConfigurationAddress(phoneNumber)

	if err != nil || resp == nil || resp.Sid == nil {

		tc.p.API.LogDebug("Configuration does not already exist", "phone_number", phoneNumber)
		return nil
	}

	err = tc.client.ConversationsV1.DeleteConfigurationAddress(*resp.Sid)
	if err != nil {
		tc.p.API.LogError("Error deleting phone number configuration", "phone_number", phoneNumber, "error", err.Error())
		return err
	}
	tc.p.API.LogDebug("Deleted phone number configuration", "phone_number", phoneNumber)
	return nil
}

func (tc *TwilioClient) AccountNumbers() ([]messaging.MessagingV1PhoneNumber, error) {

	var numbers []messaging.MessagingV1PhoneNumber
	params := &messaging.ListServiceParams{}
	resp, err := tc.client.MessagingV1.ListService(params)
	if err != nil {
		tc.p.API.LogError("Error getting messaging services", "error", err.Error())
		return nil, err
	}
	for _, service := range resp {

		mparams := &messaging.ListPhoneNumberParams{}
		r, errr := tc.client.MessagingV1.ListPhoneNumber(*service.Sid, mparams)
		if errr != nil {
			tc.p.API.LogError("Error getting phone numbers for service", "service_sid", *service.Sid, "error", errr.Error())
			return nil, errr
		}
		numbers = append(numbers, r...)

	}
	return numbers, nil
}

func (tc *TwilioClient) AccountNumbersStrings() ([]string, error) {
	var numbers []string
	nums, err := tc.AccountNumbers()
	if err != nil {
		return nil, err
	}
	for _, number := range nums {
		if number.PhoneNumber != nil {
			numbers = append(numbers, *number.PhoneNumber)
		}
	}
	return numbers, nil
}
