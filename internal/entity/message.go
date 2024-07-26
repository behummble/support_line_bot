package entity

import (
	"encoding/json"
)

type UserMessage struct {
	BotToken string
	ChatID int64
	UserID int64
	UserName string
	Payload string
	MessageID int64
	GroupChatID int64
}

type SupportMessage struct {
	BotToken string
	ChatID int64
	TopicID int
	Payload string
	MessageID int
}

func NewUserMessage(token string, chatID, userID, messageID, groupChatID int64, name, payload string) UserMessage {
	return UserMessage{
		BotToken: token,
		ChatID: chatID,
		UserID: userID,
		UserName: name,
		Payload: payload,
		MessageID: messageID,
	}
}

func NewUserMessageFromJSON(data []byte) (UserMessage, error) {
	var msg UserMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return UserMessage{}, err
	}

	return msg, err
}

func NewSupportMessage(token string, chatID int64, topicID int, Payload string) SupportMessage {
	return SupportMessage{
		BotToken: token,
		ChatID: chatID,
		TopicID: topicID,
		Payload: Payload,
	}
}

func NewSupportMessageFromJSON(data []byte) (SupportMessage, error) {
	var msg SupportMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return SupportMessage{}, err
	}

	return msg, err
}