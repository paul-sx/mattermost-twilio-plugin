package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

type webhookMessage struct {
	AccountSid     string  `json:"AccountSid"`
	EventType      string  `json:"EventType"`
	Source         string  `json:"Source"`
	ClientIdentity *string `json:"ClientIdentity,omitempty"`
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
	Index               *any      `json:"Index"`
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
	// hostname/plugins/sx.paul.mattermost.twilio/twilio/conversation
	router.HandleFunc("/twilio/conversation", p.handleTwilioConversation).Methods("POST")

	p.router = router
}

func (p *TwilioPlugin) handleTwilioConversation(w http.ResponseWriter, r *http.Request) {

	configuration := p.getConfiguration()
	/*body, err := io.ReadAll(r.Body)
	p.API.LogDebug("handleTwilioConversation", "body", string(body))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}*/
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	p.API.LogInfo("Twilio Webhook", "form", r.Form)
	accountSid := r.FormValue("AccountSid")
	if accountSid == "" || accountSid != configuration.TwilioSid {
		p.API.LogWarn("Invalid or missing AccountSid", "provided", accountSid, "expected", configuration.TwilioSid)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p.API.LogDebug("Message", "EventType", r.FormValue("EventType"))
	eventType := r.FormValue("EventType")
	switch eventType {
	case "onConversationAdded":

	case "onConversationRemoved":

	case "onConversationUpdated":

	case "onConversationStateUpdated":

	case "onMessageAdded":
		p.API.LogDebug("onMessageAdded")

		conversationSid := r.FormValue("ConversationSid")
		author := r.FormValue("Author")
		body := r.FormValue("Body")
		messageSid := r.FormValue("MessageSid")
		ChatServiceSid := r.FormValue("ChatServiceSid")

		// Handle message added logic here
		settings, err := p.getOrCreateConversationSettings(conversationSid)
		if err != nil {
			// Conversation does not have channel settings
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		p.API.LogDebug("settingscreated", "settings", settings.ChannelId)

		channel, errc := p.API.GetChannel(settings.ChannelId)
		if errc != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(errc.Error()))
			return
		}

		bot, err := p.getBot()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		p.API.LogDebug("bot", "bot", bot.UserId)

		// Add part here to deal with attachments or media if needed

		post := &model.Post{
			UserId:    bot.UserId,
			ChannelId: channel.Id,
			Message:   "<" + author + ">: " + body,
			Props: map[string]interface{}{
				"twilio_conversation_sid": conversationSid,
				"sent_by_twilio":          true,
				"twilio_message_sid":      messageSid,
			},
		}
		newpost, errp := p.API.CreatePost(post)
		if errp != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(errp.Error()))
			return
		}
		p.API.LogDebug("postcreated", "post", newpost.Id)

		if media := r.FormValue("Media"); media != "" {
			p.API.LogDebug("media", "media", media)
			var items []map[string]interface{}
			if err = json.Unmarshal([]byte(media), &items); err != nil {
				p.API.LogError("Could not unmarshal media", "error", err.Error())
				return
			}
			for _, item := range items {
				if sid, ok := item["Sid"].(string); ok {
					if Filename, ok := item["Filename"].(string); ok {
						resp, err := p.twilio.DownloadMedia(ChatServiceSid, sid)
						if err != nil {
							p.API.LogError("Could not download media", "error", err.Error())
							return
						}
						file, ferr := p.API.UploadFile(resp, channel.Id, Filename)
						if ferr != nil {
							p.API.LogError("Could not upload media", "error", ferr.Error())
							return
						}
						newpost.FileIds = append(newpost.FileIds, file.Id)
						if _, err := p.API.UpdatePost(newpost); err != nil {
							p.API.LogError("Could not update post with media", "error", err.Error())
							return
						}
					}
				}
			}

		}

	case "onMessageUpdated":

	case "onMessageRemoved":

	case "onParticipantAdded":

	case "onParticipantRemoved":

	case "onParticipantUpdated":

	case "onDeliveryUpdated":

	case "onUserAdded":

	case "onUserUpdated":

	}

	w.WriteHeader(http.StatusOK)
}

func (p *TwilioPlugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.API.LogDebug("ServeHTTP", "path", r.URL.Path)
	if p.router != nil {
		p.router.ServeHTTP(w, r)
	}
}
