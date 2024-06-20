package entity

import (
	"github.com/behummble/support_line_bot/pkg/encoding"
)

type TopicData struct {
	BotToken string
	ChatID int64
	UserID int64
	TopicID int
}

func NewTopic(token string, chatID, userID int64, topicID int) TopicData {
	return TopicData{
		BotToken: token,
		ChatID: chatID,
		UserID: userID,
		TopicID: topicID,
	}
}

func NewTopicFromJSON(data []byte) (TopicData, error) {
	res, err := encoding.FromJSON(data, TopicData{})
	if err != nil {
		return TopicData{}, err
	}
	return res.(TopicData), err
}
