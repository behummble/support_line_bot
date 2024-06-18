package supportline

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"gopkg.in/telebot.v3"
	"github.com/robfig/cron"
)

var (
	topicUserKey = "topic:user:{%d}"
	topicSupportKey = "topic:{%d}"
)

type DB interface {
	NewTopic(ctx context.Context, topicUserKey, topicSupportKey string, topicData string) error
	Topic(ctx context.Context, topic string) (string, error)
	AllTopics(ctx context.Context) ([]string, error)
	ClearTopics(ctx context.Context) error
}

type SupportService struct {
	log *slog.Logger
	bot *telebot.Bot
	db DB
	chatID int64
	chat *telebot.Chat
	cron *cron.Cron
}

type Message struct {
	BotToken []byte
	ChatID int64
	UserID int64
	UserName string
	Payload string
	MessageID int64
}

type TopicData struct {
	BotToken string
	ChatID int64
	UserID int64
	TopicID int
}

type SupportMessage struct {
	ChatID int64
	TopicID int
	Text string
}

func New(log *slog.Logger, db DB, chatID int64) *SupportService {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		panic(err)
	}

	return &SupportService{
		log: log,
		db: db,
		chatID: chatID,
		cron: cron.NewWithLocation(loc),
	}
}

func(support *SupportService) ProcessUserMessage(msg string) {
	telegramMessage, err := parseUserMessage(msg)
	if err != nil {
		support.log.Error("ParseUserMessage", err)
		return
	}
	bot, err := initBot(telegramMessage.BotToken)
	err = support.handleUserMessage(telegramMessage)
	if err != nil {
		support.log.Error("HandleMessage", err)
	}
}

func(support *SupportService) ProcessSupportMessage(msg []byte) {
	supportMsg, err := parseSupportMessage(msg)
	if err != nil {
		support.log.Error("ParseSupportMessage", err)
		return 
	}
	err = support.handleSupportMessage(supportMsg)
	if err != nil {
		support.log.Error("HandleMessage", err)
	}
}

func(support *SupportService) RemoveTopics() {
	support.cron.AddFunc("@midnight", support.clearTopicsFunc())
	support.cron.Start()
}

func(support *SupportService) handleUserMessage(telegramMessage Message) error {
	topic, err := support.db.Topic(
		context.Background(), 
		fmt.Sprintf(topicUserKey, telegramMessage.UserID))
	if err != nil {
		return err
	}

	if topic != "" {
		topicData, err := parseTopic(topic)
		if err != nil {
			return err
		}
		return support.transferMessageToTopic(topicData.TopicID, telegramMessage)
	} else {
		return support.createTopic(telegramMessage)
	}
}

func(support *SupportService) handleSupportMessage(supportMsg SupportMessage) error {
	topicInfo, err := support.db.Topic(
		context.Background(), 
		fmt.Sprintf(topicSupportKey, supportMsg.TopicID))

	if err != nil {
		return err
	}

	if topicInfo != "" {
		topicData, err := parseTopic(topicInfo)
		if err != nil {
			return err
		}
		return support.transferMessageToUser(topicData.ChatID, supportMsg.Text)
	} else {
		return fmt.Errorf("couldn't find the topic %d from the support message %s", supportMsg.TopicID, supportMsg.Text)
	}
}

func parseUserMessage(msg string) (Message, error) {
	var res Message
	err := json.Unmarshal([]byte(msg), &res)
	return res, err
}

func parseSupportMessage(msg []byte) (SupportMessage, error) {
	var res SupportMessage
	err := json.Unmarshal(msg, &res)
	return res, err
}

func (support *SupportService) transferMessageToTopic(topicID int, telegramMessage Message) error {
	opts := &telebot.SendOptions{
		ThreadID: topicID,
	}

	chat, err := support.bot.ChatByID(telegramMessage.ChatID)
	if err != nil {
		return err
	}

	msg := &telebot.Message{
		ID: int(telegramMessage.MessageID), 
		Chat: chat}

	_, err = support.bot.Forward(
		support.chat, 
		msg, 
		opts) 
	
	return err
}

func (support *SupportService) transferMessageToUser(chatID int64, payload string) error {
	_, err := support.bot.Send(telebot.ChatID(chatID), payload)
	return err
}

func (support *SupportService) createTopic(telegramMessage Message) error {
	topic, err := support.bot.CreateTopic(support.chat, generateTopic(telegramMessage.UserName))
	if err != nil {
		return err
	}

	topicData, err := prepareTopicData(telegramMessage, topic.ThreadID)
	if err != nil {
		return err
	}

	err = support.db.NewTopic(
		context.Background(),
		fmt.Sprintf(topicUserKey, telegramMessage.UserID),
		fmt.Sprintf(topicSupportKey, topic.ThreadID),
		topicData,
	)
	if err != nil {
		return err
	} else {
		return support.transferMessageToTopic(topic.ThreadID, telegramMessage)
	}
}

func generateTopic(userName string) *telebot.Topic {
	return &telebot.Topic{
			Name: userName,
			IconColor: 0,
		}
}

func prepareTopicData(msg Message, topicID int) (string, error) {
	data := TopicData{
		BotID: msg.BotID,
		ChatID: msg.ChatID,
		UserID: msg.UserID,
		TopicID: topicID,
	}

	res, err := json.Marshal(data)
	return string(res), err
}

func (support *SupportService) clearTopicsFunc() func() {
	return func() {
		support.deleteTopicsInService()
		support.deleteTopicsInDB()
	}
}

func (support *SupportService) deleteTopicsInService() {
	keys, err := support.db.AllTopics(context.Background())
	if err != nil {
		support.log.Error("GetAllTopics", err)
		return
	}
	
	for _, key := range keys {
		topic, err := support.db.Topic(
			context.Background(),
	 		key)
		if err != nil {
			support.log.Error("ExecuteIDInTopicKey", err)
			continue
		}
		topicData, err := parseTopic(topic)
		if err != nil {
			support.log.Error("ExecuteIDInTopicKey", err)
			continue
		}
		chat, err := support.bot.ChatByID(topicData.ChatID)
		if err != nil {
			support.log.Error("ParseChat", err)
			continue
		}
		teleTopic := &telebot.Topic {
			ThreadID: topicData.TopicID,
		}
		err = support.bot.CloseTopic(
			chat,
			teleTopic)
		if err != nil {
			support.log.Error("CloseTopic", err)
		}
	}
}

func (sbot *SupportService) deleteTopicsInDB() {
	err := sbot.db.ClearTopics(context.Background())
	if err != nil {
		sbot.log.Error("SheduledFlushTopics", err)
	}
}

func parseTopic(data string) (TopicData, error) {
	var topic TopicData
	err := json.Unmarshal([]byte(data), &topic)
	return topic, err
}

func (topic TopicData) ID() int {
	return topic.TopicID
}

func initBot(token []byte) {
	
}