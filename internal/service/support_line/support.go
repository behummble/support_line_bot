package supportline

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/robfig/cron"
	"gopkg.in/telebot.v3"

	"github.com/behummble/support_line_bot/internal/entity"
	"github.com/behummble/support_line_bot/internal/service/bot"
	"github.com/behummble/support_line_bot/pkg/crypto"
	"github.com/behummble/support_line_bot/pkg/encoding"
)

const (
	topicUserKey = "chatid{%d}:topic:user:{%d}"
	topicSupportKey = "chatid{%d}:topic:{%d}"
	allTopics = "topic:list"
)

type DB interface {
	NewTopic(ctx context.Context, topicUserKey, topicSupportKey, topicListKey, topicData string) error
	Topic(ctx context.Context, topic string) (string, error)
	AllTopics(ctx context.Context, keys string) ([]string, error)
	ClearTopics(ctx context.Context) error
}

type Support struct {
	log *slog.Logger
	db DB
	chatID int64
	timeout int
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
		support.log.Error("Can`t parse user message", "Error", err)
		return
	}

	bot, err := bot.New(support.log, telegramMessage.BotToken, support.timeout)

	if err != nil {
		support.log.Error("Can`t initialize bot while process user message", "Error", err)
		return
	}
	defer bot.Close()

	supportChat, err := bot.ChatByID(telegramMessage.GroupChatID)
	if err != nil {
		support.log.Error("Can`t initialize chat while process user message", "Error", err)
		return
	}

	err = support.handleUserMessage(telegramMessage, bot, supportChat)
	if err != nil {
		support.log.Error("Handle user message", "Error", err)
	}
}

func(support *Support) ProcessSupportMessage(msg []byte) {
	supportMsg, err := entity.NewSupportMessageFromJSON(msg)
	if err != nil {
		support.log.Error("Can`t parse support message", "Error", err)
		return 
	}

	bot, err := bot.New(support.log, supportMsg.BotToken, support.timeout)
	if err != nil {
		support.log.Error("Can`t initialize bot while process user message", "Error", err)
		return
	}
	defer bot.Close()

	err = support.handleSupportMessage(supportMsg, bot)
	if err != nil {
		support.log.Error("Handle support message", "Error", err)
	}
}

func(support *Support) RemoveTopics() {
	support.cron.AddFunc("@midnight", support.clearTopicsFunc())
	support.cron.Start()
}

func(support *Support) handleUserMessage(telegramMessage entity.UserMessage, bot *bot.Bot, supportChat *telebot.Chat) error {
	topic, err := support.db.Topic(
		context.Background(), 
		fmt.Sprintf(topicUserKey, telegramMessage.GroupChatID, telegramMessage.UserID))
	if err != nil {
		return err
	}

	if topic != "" {
		jsonTopic, err :=  crypto.DecryptData(topic)
		if err != nil {
			return err
		}

		topicData, err := entity.NewTopicFromJSON([]byte(jsonTopic))
		if err != nil {
			return err
		}

		return support.transferMessageToTopic(topicData.TopicID, telegramMessage, bot, supportChat)
	} else {
		return support.createTopic(telegramMessage, bot, supportChat)
	}
}

func(support *Support) handleSupportMessage(supportMsg entity.SupportMessage, bot *bot.Bot) error {
	topicInfo, err := support.db.Topic(
		context.Background(), 
		fmt.Sprintf(topicSupportKey, supportMsg.ChatID, supportMsg.TopicID))

	if err != nil {
		return err
	}

	if topicInfo != "" {
		jsonTopic, err :=  crypto.DecryptData(topicInfo)
		if err != nil {
			return err
		}

		topicData, err := entity.NewTopicFromJSON([]byte(jsonTopic))
		if err != nil {
			return err
		}
		return support.transferMessageToUser(topicData.ChatID, supportMsg.Payload, bot)
	} else {
		return fmt.Errorf("couldn't find the topic %d from the support message %s", supportMsg.TopicID, supportMsg.Payload)
	}
}

func (support *Support) transferMessageToTopic(topicID int, telegramMessage entity.UserMessage, bot *bot.Bot, supportChat *telebot.Chat) error {
	opts := &telebot.SendOptions{
		ThreadID: topicID,
	}

	userChat, err := bot.ChatByID(telegramMessage.ChatID)
	if err != nil {
		return err
	}

	msg := &telebot.Message{
		ID: int(telegramMessage.MessageID), 
		Chat: userChat}

	_, err = bot.Forward(
		supportChat, 
		msg, 
		opts)
	
	return err
}

func (support *Support) transferMessageToUser(chatID int64, payload string, bot *bot.Bot) error {
	_, err := bot.Send(telebot.ChatID(chatID), payload, &telebot.SendOptions{})
	return err
}

func (support *Support) createTopic(telegramMessage entity.UserMessage, bot *bot.Bot, supportChat *telebot.Chat) error {
	topic, err := bot.CreateTopic(supportChat, generateTopic(telegramMessage.UserName))
	if err != nil {
		return err
	}

	topicData, err := encoding.ToJSON(entity.NewTopic(
		bot.Token(),
		telegramMessage.ChatID,
		telegramMessage.UserID,
		telegramMessage.GroupChatID,
		topic.ThreadID,))
		
	if err != nil {
		return err
	}

	encryptTopicData, err := crypto.EncryptData(topicData)
	if err != nil {
		return err
	}

	err = support.db.NewTopic(
		context.Background(),
		fmt.Sprintf(topicUserKey, telegramMessage.GroupChatID, telegramMessage.UserID),
		fmt.Sprintf(topicSupportKey, telegramMessage.GroupChatID, topic.ThreadID),
		allTopics,
		encryptTopicData,
	)

	if err != nil {
		return err
	} else {
		return support.transferMessageToTopic(topic.ThreadID, telegramMessage, bot, supportChat)
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
	support.log.Info("Start delete topics")

	keys, err := support.db.AllTopics(context.Background(), allTopics)
	if err != nil {
		support.log.Error("Failed to get all topics", "Error", err)
		return
	}

	bots := make(map[string]*bot.Bot)
	groupChats := make(map[int64]*telebot.Chat)
	var waitGroup sync.WaitGroup

	support.log.Info(fmt.Sprintf("The number of topics to delete: %d", len(keys)))

	for _, key := range keys {
		waitGroup.Add(1)
		go func(key string) {
			defer waitGroup.Done()
			topic, err := support.db.Topic(
				context.Background(),
				key)
			if err != nil {
				support.log.Error("Failed to execute topic data from DB", "Error", err)
				return
			}

			jsonTopic, err :=  crypto.DecryptData(topic)
			if err != nil {
				support.log.Error("Can`t decrypt topic data", "Error", err)
				return
			}

			topicData, err := entity.NewTopicFromJSON([]byte(jsonTopic))
			if err != nil {
				support.log.Error("Can`t parse topic data from json", "Error", err)
				return
			}
			
			if _, ok := bots[topicData.BotToken]; !ok {
				bot, err := bot.NewWithoutDecryption(support.log, topicData.BotToken, support.timeout)
				if err != nil {
					support.log.Error("Can`t initialize bot in sheduling delete topics", "Error", err)
					return
				}
				bots[topicData.BotToken] = bot
			}
			bot := bots[topicData.BotToken]

			if _, ok := groupChats[topicData.GroupChatID]; !ok {
				supportChat, err := bot.ChatByID(topicData.GroupChatID)
				if err != nil {
					support.log.Error("Can`t initialize chat in sheduling delete topics", "Error", err)
					return
				}
				groupChats[topicData.GroupChatID] = supportChat
			}
			supportChat := groupChats[topicData.GroupChatID]
			
			teleTopic := &telebot.Topic {
				ThreadID: topicData.TopicID,
			}

			err = bot.DeleteTopic(
				supportChat,
				teleTopic)

			if err != nil {
				support.log.Error("Failed to delete topic", "Error", err)
			}
		} (key)	
	}

	waitGroup.Wait()
	for _, bot := range bots {
		bot.Close()
	}

	support.log.Info("Finished delete topics")
}

func (sbot *Support) deleteTopicsInDB() {
	err := sbot.db.ClearTopics(context.Background())
	if err != nil {
		sbot.log.Error("Sheduled flush topics in DB failed", "Error", err)
	}
}
