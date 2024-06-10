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
	chat *telebot.Chat
	cron *cron.Cron
}

type Message struct {
	BotID int64
	ChatID int64
	UserID int64
	UserName string
	Payload string
}

type TopicData struct {
	BotID int64
	ChatID int64
	UserID int64
	TopicID int
}

type SupportMessage struct {
	ChatID int64
	TopicID int
	Text string
}

func New(log *slog.Logger, db DB, token string, timeout int, chatID int64) *SupportService {
	bot, err := telebot.NewBot(
		telebot.Settings{
			Token: token,
			Poller: &telebot.LongPoller{Timeout: time.Second * time.Duration(timeout)},
		},
	)
	
	if err != nil {
		panic(err)
	}

	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		panic(err)
	}

	chat, err := bot.ChatByID(chatID)
	if err != nil {
		panic(err)
	}

	return &SupportService{
		log,
		bot,
		db,
		chat,
		cron.NewWithLocation(loc),
	}
}

func(sbot *SupportService) ProcessUserMessage(msg string) {
	telegramMessage, err := parseUserMessage(msg)
	if err != nil {
		sbot.log.Error("ParseUserMessage", err)
		return
	}
	err = sbot.handleUserMessage(telegramMessage)
	if err != nil {
		sbot.log.Error("HandleMessage", err)
	}
}

func(sbot *SupportService) ProcessSupportMessage(msg []byte) {
	supportMsg, err := parseSupportMessage(msg)
	if err != nil {
		sbot.log.Error("ParseSupportMessage", err)
		return 
	}
	err = sbot.handleSupportMessage(supportMsg)
	if err != nil {
		sbot.log.Error("HandleMessage", err)
	}
}

func(sbot *SupportService) RemoveTopics() {
	sbot.cron.AddFunc("@midnight", sbot.clearTopicsFunc())
	sbot.cron.Start()
}

func(sbot *SupportService) handleUserMessage(telegramMessage Message) error {
	topic, err := sbot.db.Topic(
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
		return sbot.transferMessageToTopic(topicData.TopicID, telegramMessage)
	} else {
		return sbot.createTopic(telegramMessage)
	}
}

func(sbot *SupportService) handleSupportMessage(supportMsg SupportMessage) error {
	topicInfo, err := sbot.db.Topic(
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
		return sbot.transferMessageToUser(topicData.ChatID, supportMsg.Text)
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

func (sbot *SupportService) transferMessageToTopic(topicID int, telegramMessage Message) error {
	/*msg := &telebot.Message{
		ThreadID: topicID,
		Chat: sbot.chat,
		TopicMessage: true,
	} */

	opts := &telebot.SendOptions{
		ThreadID: topicID,
	}

	_, err := sbot.bot.Send(sbot.chat, telegramMessage.Payload, opts)
	return err
}

func (sbot *SupportService) transferMessageToUser(chatID int64, payload string) error {
	_, err := sbot.bot.Send(telebot.ChatID(chatID), payload)
	return err
}

func (sbot *SupportService) createTopic(telegramMessage Message) error {

	topic, err := sbot.bot.CreateTopic(sbot.chat, generateTopic(telegramMessage.UserName))
	if err != nil {
		return err
	}

	topicData, err := prepareTopicData(telegramMessage, topic.ThreadID)
	if err != nil {
		return err
	}

	err = sbot.db.NewTopic(
		context.Background(),
		fmt.Sprintf(topicUserKey, telegramMessage.UserID),
		fmt.Sprintf(topicSupportKey, topic.ThreadID),
		topicData,
	)
	if err != nil {
		return err
	} else {
		return sbot.transferMessageToTopic(topic.ThreadID, telegramMessage)
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

func (sbot *SupportService) clearTopicsFunc() func() {
	return func() {
		sbot.deleteTopicsInService()
		sbot.deleteTopicsInDB()
	}
}

func (sbot *SupportService) deleteTopicsInService() {
	keys, err := sbot.db.AllTopics(context.Background())
	if err != nil {
		sbot.log.Error("GetAllTopics", err)
		return
	}
	
	for _, key := range keys {
		topic, err := sbot.db.Topic(
			context.Background(),
	 		key)
		if err != nil {
			sbot.log.Error("ExecuteIDInTopicKey", err)
			continue
		}
		topicData, err := parseTopic(topic)
		if err != nil {
			sbot.log.Error("ExecuteIDInTopicKey", err)
			continue
		}
		chat, err := sbot.bot.ChatByID(topicData.ChatID)
		if err != nil {
			sbot.log.Error("ParseChat", err)
			continue
		}
		teleTopic := &telebot.Topic {
			ThreadID: topicData.TopicID,
		}
		err = sbot.bot.CloseTopic(
			chat,
			teleTopic)
		if err != nil {
			sbot.log.Error("CloseTopic", err)
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