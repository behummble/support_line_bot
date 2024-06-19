package supportline

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"gopkg.in/telebot.v3"
	"github.com/robfig/cron"

	"github.com/behummble/support_line_bot/internal/service/bot"
	"github.com/behummble/support_line_bot/internal/entity"
	"github.com/behummble/support_line_bot/pkg/encoding"
)

const (
	topicUserKey = "topic:user:{%d}"
	topicSupportKey = "topic:{%d}"
)

type DB interface {
	NewTopic(ctx context.Context, topicUserKey, topicSupportKey string, topicData string) error
	Topic(ctx context.Context, topic string) (string, error)
	AllTopics(ctx context.Context) ([]string, error)
	ClearTopics(ctx context.Context) error
}

type Support struct {
	log *slog.Logger
	bot *telebot.Bot
	db DB
	chatID int64
	timeout int
	chat *telebot.Chat
	cron *cron.Cron
}

func New(log *slog.Logger, db DB, chatID int64, timeout int) *Support {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		panic(err)
	}

	return &Support{
		log: log,
		db: db,
		chatID: chatID,
		timeout: timeout,
		cron: cron.NewWithLocation(loc),
	}
}

func(support *Support) ProcessUserMessage(msg []byte) {
	telegramMessage, err := entity.NewUserMessageFromJSON(msg)
	if err != nil {
		support.log.Error("ParseUserMessage", err)
		return
	}

	bot, err := bot.New(support.log, telegramMessage.BotToken, support.timeout)
	
	err = support.handleUserMessage(telegramMessage)
	if err != nil {
		support.log.Error("HandleMessage", err)
	}
}

func(support *Support) ProcessSupportMessage(msg []byte) {
	supportMsg, err := entity.NewSupportMessageFromJSON(msg)
	if err != nil {
		support.log.Error("ParseSupportMessage", err)
		return 
	}

	err = support.handleSupportMessage(supportMsg)
	if err != nil {
		support.log.Error("HandleMessage", err)
	}
}

func(support *Support) RemoveTopics() {
	support.cron.AddFunc("@midnight", support.clearTopicsFunc())
	support.cron.Start()
}

func(support *Support) handleUserMessage(telegramMessage entity.UserMessage) error {
	topic, err := support.db.Topic(
		context.Background(), 
		fmt.Sprintf(topicUserKey, telegramMessage.UserID))
	if err != nil {
		return err
	}

	if topic != "" {
		topicData, err := entity.NewTopicFromJSON([]byte(topic))
		if err != nil {
			return err
		}
		return support.transferMessageToTopic(topicData.TopicID, telegramMessage)
	} else {
		return support.createTopic(telegramMessage)
	}
}

func(support *Support) handleSupportMessage(supportMsg entity.SupportMessage) error {
	topicInfo, err := support.db.Topic(
		context.Background(), 
		fmt.Sprintf(topicSupportKey, supportMsg.TopicID))

	if err != nil {
		return err
	}

	if topicInfo != "" {
		topicData, err := entity.NewTopicFromJSON([]byte(topicInfo))
		if err != nil {
			return err
		}
		return support.transferMessageToUser(topicData.ChatID, supportMsg.Payload)
	} else {
		return fmt.Errorf("couldn't find the topic %d from the support message %s", supportMsg.TopicID, supportMsg.Text)
	}
}

func (support *Support) transferMessageToTopic(topicID int, telegramMessage entity.UserMessage) error {
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

func (support *Support) transferMessageToUser(chatID int64, payload string) error {
	_, err := support.bot.Send(telebot.ChatID(chatID), payload)
	return err
}

func (support *Support) createTopic(telegramMessage entity.UserMessage) error {
	topic, err := support.bot.CreateTopic(support.chat, generateTopic(telegramMessage.UserName))
	if err != nil {
		return err
	}

	topicData, err := encoding.ToJSON(entity.NewTopic(
		telegramMessage.BotToken,
		telegramMessage.ChatID,
		telegramMessage.UserID,
		topic.ThreadID))
	
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

func (support *Support) clearTopicsFunc() func() {
	return func() {
		support.deleteTopicsInService()
		support.deleteTopicsInDB()
	}
}

func (support *Support) deleteTopicsInService() {
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
		topicData, err := entity.NewTopicFromJSON([]byte(topic))
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

func (sbot *Support) deleteTopicsInDB() {
	err := sbot.db.ClearTopics(context.Background())
	if err != nil {
		sbot.log.Error("SheduledFlushTopics", err)
	}
}

func initBot(token []byte) {
	
}