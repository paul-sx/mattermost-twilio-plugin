package main

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

type webhookMessage struct {
	AccountSid     string `json:"AccountSid"`
	EventType      string `json:"EventType"`
	Source         string `json:"Source"`
	ClientIdentity string `json:"ClientIdentity"`
}

type webhookOnConversationAdded struct {
	webhookMessage
	ConversationSid     string    `json:"ConversationSid"`
	DateCreated         time.Time `json:"DateCreated"`
	DateUpdated         time.Time `json:"DateUpdated"`
	FriendlyName        *string   `json:"FriendlyName,omitempty"`
	UniqueName          *string   `json:"UniqueName,omitempty"`
	Attributes          string    `json:"Attributes"`
	ChatServiceSid      string    `json:"ChatServiceSid"`
	MessagingServiceSid string    `json:"MessagingServiceSid"`
	MessagingBinding    struct {
		ProxyAddress     *string `json:"ProxyAddress,omitempty"`
		Address          *string `json:"Address,omitempty"`
		ProjectedAddress *string `json:"ProjectedAddress,omitempty"`
		AuthorAddress    *string `json:"AuthorAddress,omitempty"`
	} `json:"MessagingBinding"`
	State string `json:"State"`
}

type webhookOnConversationRemoved struct {
	webhookMessage
	ConversationSid     string    `json:"ConversationSid"`
	DateCreated         time.Time `json:"DateCreated"`
	DateUpdated         time.Time `json:"DateUpdated"`
	DateRemoved         time.Time `json:"DateRemoved"`
	FriendlyName        *string   `json:"FriendlyName,omitempty"`
	UniqueName          *string   `json:"UniqueName,omitempty"`
	Attributes          string    `json:"Attributes"`
	ChatServiceSid      string    `json:"ChatServiceSid"`
	MessagingServiceSid string    `json:"MessagingServiceSid"`
	State               string    `json:"State"`
}

type webhookOnConversationUpdated struct {
	webhookMessage
	ConversationSid     string    `json:"ConversationSid"`
	DateCreated         time.Time `json:"DateCreated"`
	DateUpdated         time.Time `json:"DateUpdated"`
	FriendlyName        *string   `json:"FriendlyName,omitempty"`
	UniqueName          *string   `json:"UniqueName,omitempty"`
	Attributes          string    `json:"Attributes"`
	ChatServiceSid      string    `json:"ChatServiceSid"`
	MessagingServiceSid string    `json:"MessagingServiceSid"`
	State               string    `json:"State"`
}
type webhookOnConversationStateUpdated struct {
	webhookMessage
	ChatServiceSid      string    `json:"ChatServiceSid"`
	StateUpdated        time.Time `json:"StateUpdated"`
	StateFrom           string    `json:"StateFrom"`
	StateTo             string    `json:"StateTo"`
	ConversationSid     string    `json:"ConversationSid"`
	Reason              string    `json:"Reason"`
	MessagingServiceSid string    `json:"MessagingServiceSid"`
}

type webhookOnMessageAdded struct {
	webhookMessage
	ConversationSid     string    `json:"ConversationSid"`
	MessageSid          string    `json:"MessageSid"`
	MessagingServiceSid string    `json:"MessagingServiceSid"`
	Index               int       `json:"Index"`
	DateCreated         time.Time `json:"DateCreated"`
	Body                string    `json:"Body"`
	Author              string    `json:"Author"`
	ParticipantSid      *string   `json:"ParticipantSid,omitempty"`
	Attributes          string    `json:"Attributes"`
	Media               *string   `json:"Media,omitempty"`
}

type webhookOnMessageUpdated struct {
	webhookMessage
	ConversationSid string    `json:"ConversationSid"`
	MessageSid      string    `json:"MessageSid"`
	Index           int       `json:"Index"`
	DateCreated     time.Time `json:"DateCreated"`
	DateUpdated     time.Time `json:"DateUpdated"`
	Body            string    `json:"Body"`
	Author          string    `json:"Author"`
	ParticipantSid  *string   `json:"ParticipantSid,omitempty"`
	Attributes      string    `json:"Attributes"`
	Media           *string   `json:"Media,omitempty"`
}

type webhookOnMessageRemoved struct {
	webhookMessage
	ConversationSid string    `json:"ConversationSid"`
	MessageSid      string    `json:"MessageSid"`
	Index           int       `json:"Index"`
	DateCreated     time.Time `json:"DateCreated"`
	DateUpdated     time.Time `json:"DateUpdated"`
	DateRemoved     time.Time `json:"DateRemoved"`
	Body            string    `json:"Body"`
	Author          string    `json:"Author"`
	ParticipantSid  *string   `json:"ParticipantSid,omitempty"`
	Attributes      string    `json:"Attributes"`
	Media           *string   `json:"Media,omitempty"`
}

type webhookOnParticipantAdded struct {
	webhookMessage
	ConversationSid  string    `json:"ConversationSid"`
	ParticipantSid   string    `json:"ParticipantSid"`
	DateCreated      time.Time `json:"DateCreated"`
	Identity         *string   `json:"Identity,omitempty"`
	RoleSid          string    `json:"RoleSid"`
	Attributes       string    `json:"Attributes"`
	MessagingBinding struct {
		ProxyAddress     *string `json:"ProxyAddress,omitempty"`
		Address          *string `json:"Address,omitempty"`
		ProjectedAddress *string `json:"ProjectedAddress,omitempty"`
		Type             string  `json:"Type"`
	} `json:"MessagingBinding"`
}

type webhookOnParticipantRemoved struct {
	webhookMessage
	ConversationSid  string    `json:"ConversationSid"`
	ParticipantSid   string    `json:"ParticipantSid"`
	DateCreated      time.Time `json:"DateCreated"`
	DateUpdated      time.Time `json:"DateUpdated"`
	DateRemoved      time.Time `json:"DateRemoved"`
	Identity         *string   `json:"Identity,omitempty"`
	RoleSid          string    `json:"RoleSid"`
	Attributes       string    `json:"Attributes"`
	MessagingBinding struct {
		ProxyAddress     *string `json:"ProxyAddress,omitempty"`
		Address          *string `json:"Address,omitempty"`
		ProjectedAddress *string `json:"ProjectedAddress,omitempty"`
		Type             string  `json:"Type"`
	} `json:"MessagingBinding"`
}
type webhookOnParticipantUpdated struct {
	webhookMessage
	ConversationSid  string    `json:"ConversationSid"`
	ParticipantSid   string    `json:"ParticipantSid"`
	DateCreated      time.Time `json:"DateCreated"`
	DateUpdated      time.Time `json:"DateUpdated"`
	Identity         *string   `json:"Identity,omitempty"`
	RoleSid          string    `json:"RoleSid"`
	Attributes       string    `json:"Attributes"`
	MessagingBinding struct {
		ProxyAddress     *string `json:"ProxyAddress,omitempty"`
		Address          *string `json:"Address,omitempty"`
		ProjectedAddress *string `json:"ProjectedAddress,omitempty"`
		Type             string  `json:"Type"`
	} `json:"MessagingBinding"`
	LastReadMessageIndex int `json:"LastReadMessageIndex"`
}

type webhookOnDeliveryUpdated struct {
	webhookMessage
	AccountSid           string    `json:"AccountSid"`
	ConversationSid      string    `json:"ConversationSid"`
	ChatServiceSid       string    `json:"ChatServiceSid"`
	MessageSid           string    `json:"MessageSid"`
	DeliveryRecipientSid string    `json:"DeliveryRecipientSid"`
	ChannelMessageSid    string    `json:"ChannelMessageSid"`
	ParticipantSid       string    `json:"ParticipantSid"`
	Status               string    `json:"Status"`
	ErrorCode            int       `json:"ErrorCode"`
	DateCreated          time.Time `json:"DateCreated"`
	DateUpdated          time.Time `json:"DateUpdated"`
}

type webhookOnUserAdded struct {
	webhookMessage
	ChatServiceSid string    `json:"ChatServiceSid"`
	UserSid        string    `json:"UserSid"`
	DateCreated    time.Time `json:"DateCreated"`
	Identity       *string   `json:"Identity,omitempty"`
	RoleSid        string    `json:"RoleSid"`
	Attributes     string    `json:"Attributes"`
	FriendlyName   string    `json:"FriendlyName"`
}
type webhookOnUserUpdated struct {
	webhookMessage
	ChatServiceSid string    `json:"ChatServiceSid"`
	UserSid        string    `json:"UserSid"`
	DateCreated    time.Time `json:"DateCreated"`
	DateUpdated    time.Time `json:"DateUpdated"`
	Identity       *string   `json:"Identity,omitempty"`
	RoleSid        string    `json:"RoleSid"`
	Attributes     string    `json:"Attributes"`
	FriendlyName   string    `json:"FriendlyName"`
	IsOnline       bool      `json:"isOnline"`
	IsNotifiable   bool      `json:"isNotifiable"`
}

func (p *TwilioPlugin) initializeRouter() {
	router := mux.NewRouter()

	router.HandleFunc("/twilio/conversation", p.handleTwilioConversation).Methods("POST")

	p.router = router
}

func (p *TwilioPlugin) handleTwilioConversation(w http.ResponseWriter, r *http.Request) {
	var message map[string]interface{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &message); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if eventType, ok := message["EventType"].(string); ok {
		switch eventType {
		case "onConversationAdded":
			var conversationAdded webhookOnConversationAdded
			if err := json.Unmarshal(body, &conversationAdded); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// Handle conversation added logic here

		case "onConversationRemoved":
			var conversationRemoved webhookOnConversationRemoved
			if err := json.Unmarshal(body, &conversationRemoved); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// Handle conversation removed logic here

		case "onConversationUpdated":
			var conversationUpdated webhookOnConversationUpdated
			if err := json.Unmarshal(body, &conversationUpdated); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// Handle conversation updated logic here

		case "onConversationStateUpdated":
			var conversationStateUpdated webhookOnConversationStateUpdated
			if err := json.Unmarshal(body, &conversationStateUpdated); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// Handle conversation state updated logic here

		case "onMessageAdded":
			var messageAdded webhookOnMessageAdded
			if err := json.Unmarshal(body, &messageAdded); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// Handle message added logic here
			settings, err := p.getOrCreateConversationSettings(messageAdded.ConversationSid)
			if err != nil {
				// Conversation does not have channel settings
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			channel, err := p.API.GetChannel(settings.ChannelId)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			bot, err := p.getBot()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Add part here to deal with attachments or media if needed

			post := &model.Post{
				UserId:    bot.UserId,
				ChannelId: channel.Id,
				Message:   messageAdded.Body,
				Props: map[string]interface{}{
					"twilio_conversation_sid": messageAdded.ConversationSid,
					"sent_by_twilio":          true,
				},
			}
			if _, err := p.API.CreatePost(post); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		case "onMessageUpdated":
			var messageUpdated webhookOnMessageUpdated
			if err := json.Unmarshal(body, &messageUpdated); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// Handle message updated logic here

		case "onMessageRemoved":
			var messageRemoved webhookOnMessageRemoved
			if err := json.Unmarshal(body, &messageRemoved); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// Handle message removed logic here

		case "onParticipantAdded":
			var participantAdded webhookOnParticipantAdded
			if err := json.Unmarshal(body, &participantAdded); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// Handle participant added logic here

		case "onParticipantRemoved":
			var participantRemoved webhookOnParticipantRemoved
			if err := json.Unmarshal(body, &participantRemoved); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// Handle participant removed logic here

		case "onParticipantUpdated":
			var participantUpdated webhookOnParticipantUpdated
			if err := json.Unmarshal(body, &participantUpdated); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// Handle participant updated logic here

		case "onDeliveryUpdated":
			var deliveryUpdated webhookOnDeliveryUpdated
			if err := json.Unmarshal(body, &deliveryUpdated); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// Handle delivery updated logic here

		case "onUserAdded":
			var userAdded webhookOnUserAdded
			if err := json.Unmarshal(body, &userAdded); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// Handle user added logic here

		case "onUserUpdated":
			var userUpdated webhookOnUserUpdated
			if err := json.Unmarshal(body, &userUpdated); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// Handle user updated logic here

		}
	}

	w.WriteHeader(http.StatusOK)
}

func (p *TwilioPlugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	if p.router != nil {
		p.router.ServeHTTP(w, r)
	}
}
