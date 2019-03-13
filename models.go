package bots

import "net/http"

type Bot interface {
	http.Handler
	Send(activity *Activity) (*Identification, error)
	Update(activity *Activity) (*Identification, error)
	Delete(activity *Activity) error
	GetUpdatesChannel() (<-chan *Activity, error)
	GetFile(attachment *Attachment, activity *Activity) (*http.Response, error)
	GetChannels() []string
}

type Identification struct {
	Id string `json:"id,omitempty"`
}

type ActivityType string
type CardActionType string
type TextFormat string

const (
	TypeContactRelationUpdate = ActivityType("contactRelationUpdate")
	TypeConversationUpdate    = ActivityType("conversationUpdate")
	TypeDeleteUserData        = ActivityType("deleteUserData")
	TypeMessage               = ActivityType("message")
	TypeTyping                = ActivityType("typing")
	TypeEvent                 = ActivityType("event")
	TypeEndOfConversation     = ActivityType("endOfConversation")

	TypeOpenUrl      = CardActionType("openUrl")
	TypeImBack       = CardActionType("imBack")
	TypePostBack     = CardActionType("postBack")
	TypeCall         = CardActionType("call")
	TypePlayAudio    = CardActionType("playAudio")
	TypePlayVideo    = CardActionType("playVideo")
	TypeShowImage    = CardActionType("showImage")
	TypeDownloadFile = CardActionType("downloadFile")
	TypeSignin       = CardActionType("signin")

	Markdown = TextFormat("markdown")
	Plain    = TextFormat("plain")
	Xml      = TextFormat("xml")

	ChannelSkype    = "skype"
	ChannelViber    = "viber"
	ChannelTelegram = "telegram"
	ChannelKik      = "kik"
	ChannelLine     = "line"
	ChannelWebChat  = "webchat"
	ChannelFacebook = "facebook"

	TypeHeroCard = "application/vnd.microsoft.card.hero"
)

type Activity struct {
	Identification
	Text             string               `json:"text,omitempty"`
	From             *ChannelAccount      `json:"from"`
	Conversation     *ConversationAccount `json:"conversation,omitempty"`
	ServiceUrl       string               `json:"serviceUrl"`
	ChannelData      interface{}          `json:"channelData,omitempty"`
	ChannelId        string               `json:"channelId,omitempty"`
	Recipient        *ChannelAccount      `json:"recipient"`
	Type             ActivityType         `json:"type"`
	InputHint        string               `json:"inputHint,omitempty"`
	Attachments      []*Attachment        `json:"attachments,omitempty"`
	TextFormat       TextFormat           `json:"textFormat,omitempty"`
	ReplyToId        string               `json:"replyToId,omitempty"`
	SuggestedActions *SuggestedActions    `json:"suggestedActions,omitempty"`
}

func (a *Activity) Response(message string) *Activity {
	response := Activity{}
	response.Type = "message"
	response.From = a.Recipient
	response.Recipient = a.From
	response.Text = message
	response.ReplyToId = a.Id
	response.Conversation = a.Conversation
	response.ServiceUrl = a.ServiceUrl
	response.ChannelId = a.ChannelId
	return &response
}

type ChannelAccount struct {
	Identification
	Name string `json:"name"`
}

type ConversationAccount struct {
	ChannelAccount
	IsGroup bool `json:"isGroup"`
}

type Attachment struct {
	ContentType  string      `json:"contentType,omitempty"`
	ContentUrl   string      `json:"contentUrl,omitempty"`
	ThumbnailUrl string      `json:"thumbnailUrl,omitempty"`
	Name         string      `json:"name,omitempty"`
	Content      interface{} `json:"content,omitempty"`
}

type OAuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

type CardAction struct {
	Image string         `json:"image,omitempty"`
	Text  string         `json:"text,omitempty"`
	Title string         `json:"title,omitempty"`
	Type  CardActionType `json:"type"`
	Value string         `json:"value"`
}

type CardImage struct {
	Alt string      `json:"alt,omitempty"`
	Tap *CardAction `json:"tap,omitempty"`
	Url string      `json:"url,omitempty"`
}

type HeroCard struct {
	Images  []*CardImage  `json:"images,omitempty"`
	Buttons []*CardAction `json:"buttons,omitempty"`
	Title   string        `json:"title,omitempty"`
	Text    string        `json:"text,omitempty"`
}

type SuggestedActions struct {
	Actions []*CardAction `json:"actions"`
	To      []string      `json:"to,omitempty"`
}
