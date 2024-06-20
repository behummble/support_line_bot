package entity

import (
	"github.com/behummble/support_line_bot/pkg/encoding"
)

type UserMessage struct {
	BotToken []byte
	ChatID int64
	UserID int64
	UserName string
	Payload string
	MessageID int64
}

type SupportMessage struct {
	BotToken []byte
	ChatID int64
	TopicID int
	Payload string
}

func NewUserMessage(token []byte, chatID, userID, messageID int64, name, payload string) UserMessage {
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
	res, err := encoding.FromJSON(data, UserMessage{})
	if err != nil {
		return UserMessage{}, err
	}
	return res.(UserMessage), err
}

func NewSupportMessage(chatID int64, topicID int, Payload string) SupportMessage {
	return SupportMessage{
		ChatID: chatID,
		TopicID: topicID,
		Payload: Payload,
	}
}

func NewSupportMessageFromJSON(data []byte) (SupportMessage, error) {
	res, err := encoding.FromJSON(data, SupportMessage{})
	if err != nil {
		return SupportMessage{}, err
	}
	return res.(SupportMessage), err
}