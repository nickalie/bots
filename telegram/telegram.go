package telegram

type SendMessage struct {
	ChatID      int64       `json:"chat_id"`
	Text        string      `json:"text"`
	ReplyMarkup interface{} `json:"reply_markup,omitempty"`
}

func (s *SendMessage) ToChannelData() map[string]interface{} {
	result := make(map[string]interface{})
	result["method"] = "sendMessage"
	result["parameters"] = s
	return result
}
