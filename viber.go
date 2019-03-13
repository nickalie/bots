package bots

import (
	"errors"
	"fmt"
	"github.com/nickalie/viber"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var viberChannels = []string{ChannelViber}

type ViberBot struct {
	bot     *viber.Viber
	updates chan *Activity
	config  *ViberBotConfig
}

type ViberBotConfig struct {
	Token               string
	WebHookURL          string
	ConversationStarted func(m *Activity) *Activity
}

func NewViberBot(config *ViberBotConfig) (*ViberBot, error) {
	result := &ViberBot{updates: make(chan *Activity), config: config}
	result.bot = &viber.Viber{
		AppKey: config.Token,
		Sender: viber.Sender{
			Name: "To Kindle Bot",
		},
		Message: result.messageHandler,
	}

	if config.ConversationStarted != nil {
		result.bot.ConversationStarted = result.conversationStartedHandler
	}

	go func() {
		_, err := result.bot.SetWebhook(config.WebHookURL, nil)

		if err != nil {
			fmt.Printf("unabled to set viber webhook: %v\n", err)
			os.Exit(1)
		}
	}()

	return result, nil
}

func (b *ViberBot) conversationStartedHandler(v *viber.Viber, u viber.User, conversationType, context string, subscribed bool, token uint64, t time.Time) viber.Message {
	m := b.config.ConversationStarted(viberToActivity(b.bot.NewTextMessage(""), &u))
	return b.activityToViber(m)
}

func (b *ViberBot) messageHandler(v *viber.Viber, u viber.User, m viber.Message, token uint64, t time.Time) {
	switch v := m.(type) {
	case *viber.TextMessage:
		b.updates <- viberToActivity(v, &u)
	case *viber.FileMessage:
		m := viberToActivity(&v.TextMessage, &u)
		a := &Attachment{
			Name:       v.FileName,
			ContentUrl: v.Media,
		}
		m.Attachments = append(m.Attachments, a)

		b.updates <- m
	}
}

func (b *ViberBot) GetUpdatesChannel() (<-chan *Activity, error) {
	_, err := b.bot.SetWebhook(b.config.WebHookURL, nil)
	return b.updates, err
}

func (b *ViberBot) Send(a *Activity) (*Identification, error) {
	m := b.activityToViber(a)
	token, err := b.bot.SendMessage(a.Recipient.Id, m)

	if err != nil {
		return nil, err
	} else {
		return &Identification{Id: strconv.FormatUint(token, 10)}, nil
	}
}

func (b *ViberBot) Update(a *Activity) (*Identification, error) {
	return nil, errors.New("update isn't implemented for viber")
}

func (b *ViberBot) Delete(a *Activity) error {
	return errors.New("delete isn't implemented for viber")
}

func (b *ViberBot) GetFile(attachment *Attachment, activity *Activity) (*http.Response, error) {
	timeout := time.Duration(5 * time.Second)

	client := http.Client{
		Timeout: timeout,
	}

	return client.Get(attachment.ContentUrl)
}

func (b *ViberBot) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.bot.ServeHTTP(w, r)
}

func (b *ViberBot) GetChannels() []string {
	return viberChannels
}

func viberToActivity(m *viber.TextMessage, u *viber.User) *Activity {
	result := &Activity{}
	result.ChannelId = ChannelViber
	result.Text = m.Text
	result.From = &ChannelAccount{
		Identification: Identification{Id: u.ID},
		Name:           u.Name,
	}
	result.Conversation = &ConversationAccount{
		ChannelAccount: *result.From,
		IsGroup:        false,
	}
	result.Type = TypeMessage
	return result
}

func (b *ViberBot) activityToViber(v *Activity) viber.Message {
	if v.TextFormat == Markdown {
		v.Text = strings.Replace(v.Text, "*", "", -1)
		v.Text = strings.Replace(v.Text, "[", "", -1)
		v.Text = strings.Replace(v.Text, "]", " ", -1)
		v.Text = strings.Replace(v.Text, "<br/>", "\n", -1)
	}

	var m viber.Message
	var attachment *Attachment

	if len(v.Attachments) > 0 {
		attachment = v.Attachments[0]

		if strings.HasPrefix(attachment.ContentType, "image") {
			m = b.bot.NewPictureMessage(v.Text, attachment.ContentUrl, "")
		}
	}

	if m == nil {
		m = b.bot.NewTextMessage(v.Text)
	}

	if v.SuggestedActions != nil {
		m.SetKeyboard(suggestedActionsToViber(v.SuggestedActions.Actions))
	} else if attachment != nil && attachment.ContentType == TypeHeroCard {
		card, ok := attachment.Content.(*HeroCard)

		if ok {
			m.SetKeyboard(suggestedActionsToViber(card.Buttons))
		}
	}

	return m
}

func suggestedActionsToViber(actions []*CardAction) *viber.Keyboard {
	keyboard := &viber.Keyboard{
		Type:          "keyboard",
		DefaultHeight: false,
	}

	columns := 6 / len(actions)

	for _, action := range actions {
		vButton := viber.Button{
			Text:       action.Title,
			BgColor:    "#f6f7f9",
			Columns:    columns,
			Rows:       1,
			ActionBody: action.Value,
		}

		if action.Type == TypeOpenUrl {
			vButton.ActionType = viber.OpenURL
		} else {
			vButton.ActionType = viber.Reply
		}

		keyboard.Buttons = append(keyboard.Buttons, vButton)
	}

	return keyboard
}
