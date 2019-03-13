package bots

import (
	"errors"
	"github.com/thoas/go-funk"
	"net/http"
)

type MultiBot struct {
	bots    []Bot
	updates chan *Activity
}

func NewMultiBot(bots ...Bot) *MultiBot {
	return &MultiBot{bots: bots, updates: make(chan *Activity)}
}

func (b *MultiBot) GetUpdatesChannel() (<-chan *Activity, error) {
	for _, bot := range b.bots {
		updates, err := bot.GetUpdatesChannel()

		if err != nil {
			return nil, err
		}

		go b.startUpdates(updates)
	}

	return b.updates, nil
}

func (b *MultiBot) GetPlatforms() (result []string) {
	for _, bot := range b.bots {
		result = append(result, bot.GetChannels()...)
	}

	return
}

func (b *MultiBot) Send(activity *Activity) (*Identification, error) {
	if bot := b.findBotByChannel(activity.ChannelId); bot != nil {
		return bot.Send(activity)
	}

	return nil, errors.New("MultiBot.Send: Unknown platform: " + activity.ChannelId)
}

func (b *MultiBot) Update(activity *Activity) (*Identification, error) {
	if bot := b.findBotByChannel(activity.ChannelId); bot != nil {
		return bot.Update(activity)
	}

	return nil, errors.New("MultiBot.Send: Unknown platform: " + activity.ChannelId)
}

func (b *MultiBot) Delete(activity *Activity) error {
	if bot := b.findBotByChannel(activity.ChannelId); bot != nil {
		return bot.Delete(activity)
	}

	return errors.New("MultiBot.Send: Unknown platform: " + activity.ChannelId)
}

func (b *MultiBot) GetFile(file *Attachment, activity *Activity) (*http.Response, error) {
	if bot := b.findBotByChannel(activity.ChannelId); bot != nil {
		return bot.GetFile(file, activity)
	}

	return nil, errors.New("MultiBot.Send: Unknown platform: " + activity.ChannelId)
}

func (b *MultiBot) GetChannels() (result []string) {
	for _, bot := range b.bots {
		result = append(result, bot.GetChannels()...)
	}

	return
}

func (b *MultiBot) startUpdates(updates <-chan *Activity) {
	for {
		m := <-updates
		b.updates <- m
	}
}

func (b *MultiBot) findBotByChannel(channel string) Bot {
	for _, bot := range b.bots {
		if funk.Contains(bot.GetChannels(), channel) {
			return bot
		}
	}

	return nil
}

func (b *MultiBot) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
