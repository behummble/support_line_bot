package entity

import (
	//"github.com/behummble/support_line_bot/pkg/encoding"
	"encoding/json"
)

type UserMessage struct {
	BotToken string
	ChatID int64
	UserID int64
	UserName string
	Payload string
	MessageID int64
}

type SupportMessage struct {
	BotToken string
	ChatID int64
	TopicID int
	Payload string
}

func NewUserMessage(token string, chatID, userID, messageID int64, name, payload string) UserMessage {
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
	//res, err := encoding.FromJSON(data, msg)
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return UserMessage{}, err
	}
	//return res.(UserMessage), err
	return msg, err
}

func NewSupportMessage(chatID int64, topicID int, Payload string) SupportMessage {
	return SupportMessage{
		ChatID: chatID,
		TopicID: topicID,
		Payload: Payload,
	}
}

func NewSupportMessageFromJSON(data []byte) (SupportMessage, error) {
	//res, err := encoding.FromJSON(data, SupportMessage{})
	var msg SupportMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return SupportMessage{}, err
	}
	//return res.(SupportMessage), err
	return msg, err
}