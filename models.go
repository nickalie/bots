package msbot

import (

)

type Identification struct {
	Id string `json:"id,omitempty"`
}

type Activity struct {
	Identification
	Text         string               `json:"text,omitempty"`
	From         *ChannelAccount      `json:"from"`
	Conversation *ConversationAccount `json:"conversation,omitempty"`
	ServiceUrl   string               `json:"serviceUrl"`
	ChannelData  interface{}          `json:"channelData,omitempty"`
	ChannelId    string               `json:"channelId,omitempty"`
	Recipient    *ChannelAccount      `json:"recipient"`
	Type         string               `json:"type"`
	InputHint    string               `json:"inputHint,omitempty"`
	Attachments  []*Attachment        `json:"attachments,omitempty"`
	TextFormat   string               `json:"textFormat,omitempty"`
	ReplyToId    string               `json:"replyToId,omitempty"`
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
	ContentType  string      `json:"contentType"`
	ContentUrl   string      `json:"contentUrl"`
	ThumbnailUrl string      `json:"thumbnailUrl"`
	Name         string      `json:"name"`
	Content      interface{} `json:"content,omitempty"`
}

type OAuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

type CardAction struct {
	Image string `json:"image,omitempty"`
	Text  string `json:"text,omitempty"`
	Title string `json:"title,omitempty"`
}

type CardImage struct {
	Alt string      `json:"alt,omitempty"`
	Tap *CardAction `json:"tap,omitempty"`
	Url string      `json:"url,omitempty"`
}

type HeroCard struct {
	Images []*CardImage `json:"images,omitempty"`
}