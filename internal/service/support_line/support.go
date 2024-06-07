package supportline

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"
	"gopkg.in/telebot.v3"
	"github.com/robfig/cron"
)

var (
	topicKey = "topic:{%d}"
)

type DB interface {
	NewTopic(ctx context.Context, topic string, topicData string) error
	Topic(ctx context.Context, topic string) (string, error)
	AllTopics(ctx context.Context) ([]string, error)
	ClearTopics(ctx context.Context) error
}

type SupportService struct {
	log *slog.Logger
	bot *telebot.Bot
	db DB
	chatID int64
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
	TopicID int64
}

type SupportData struct {
	ChatID int64
	TopicID int64
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

	return &SupportService{
		log,
		bot,
		db,
		chatID,
		cron.NewWithLocation(loc),
	}
}

func(sbot *SupportService) ProcessMessage(msg string) {
	telegramMessage, err := parseMessage(msg)
	if err != nil {
		return
	}
	err = sbot.handleMessage(telegramMessage)
	if err != nil {
		sbot.log.Error("HandleMessage", err)
	}
}

func(sbot *SupportService) RemoveTopics() {
	sbot.cron.AddFunc("@midnight", sbot.clearTopicsFunc())
	sbot.cron.Start()
}

func(sbot *SupportService) handleMessage(telegramMessage Message) error {
	topic, err := sbot.db.Topic(
		context.Background(), 
		fmt.Sprintf(topicKey, telegramMessage.UserID))
	if err != nil {
		return err
	}

	if topic != "" {
		
		return sbot.transferMessageToTopic(topic, telegramMessage)
	} else {
		return sbot.createTopic(telegramMessage)
	}
}

func parseMessage(msg string) (Message, error) {
	var res Message
	err := json.Unmarshal([]byte(msg), &res)
	return res, err
}

func (sbot *SupportService) transferMessageToTopic(topicID int, telegramMessage Message) error {
	
	msg := &telebot.Message{
		ThreadID: int(id),
		Chat: &telebot.Chat{ID: telegramMessage.ChatID},
	}

	_, err = sbot.bot.Reply(msg, telegramMessage.Payload)
	return err	
}

func (sbot *SupportService) createTopic(telegramMessage Message) error {
	chat := &telebot.Chat{
		ID: sbot.chatID,
		Private: false,
	}

	topic, err := sbot.bot.CreateTopic(chat, generateTopic(telegramMessage.UserName))
	if err != nil {
		return err
	}

	topicData, err := prepareTopicData(telegramMessage, int64(topic.ThreadID))
	if err != nil {
		return err
	}

	err = sbot.db.NewTopic(
		context.Background(),
		fmt.Sprintf(topicKey, telegramMessage.UserID),
		topicData,
	)
	if err != nil {
		return err
	} else {
		return sbot.transferMessageToTopic(topic.)
	}
}

func generateTopic(userName string) *telebot.Topic {
	return &telebot.Topic{
			Name: userName,
			IconColor: 0,
		}
}

func prepareTopicData(msg Message, topicID int64) (string, error) {
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
	topics, err := sbot.db.AllTopics(context.Background())
	if err != nil {
		sbot.log.Error("GetAllTopics", err)
		return
	}
	
	for _, topic := range topics {
		topicData, err := parseTopic(topic)
		if err != nil {
			sbot.log.Error("ExecuteIDInTopicKey", err)
			continue
		}
		chat := &telebot.Chat{
			ID: topicData.ChatID,
		}
		teleTopic := &telebot.Topic {
			ThreadID: int(topicData.TopicID),
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