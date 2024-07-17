package entity

import (
	"encoding/json"
)

type TopicData struct {
	BotToken string
	ChatID int64
	UserID int64
	TopicID int
	GroupChatID int64
}

func NewTopic(token string, chatID, userID, groupChatID int64, topicID int) TopicData {
	return TopicData{
		BotToken: token,
		ChatID: chatID,
		UserID: userID,
		TopicID: topicID,
		GroupChatID: groupChatID,
	}
}

func NewTopicFromJSON(data []byte) (TopicData, error) {
	var topic TopicData
	err := json.Unmarshal(data, &topic)
	if err != nil {
		return TopicData{}, err
	}

	return topic, err
}
