package supportline

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"gopkg.in/telebot.v3"
)

const (
	errorMessage = "Sorry, something went wrong, try to repeat the request later"
	startMessage = "Hi, %s"
	cantCreateTopic = "К сожалению, не смогли адресовать сообщение в службу поддержки. Попробуйте обратиться по команде ниже"
	sendMessageToSupport = "Передал ваше сообщение в службу поддержки, они ответят в ближайшее время"
)

const (
	commandNotFound = "Can`t find command %s"
)

var (
	inLine = &telebot.ReplyMarkup{ResizeKeyboard: true,}
	btnHelp = inLine.Text("✍ Связаться с поддержкой")
)

type DB interface {
	NewTopic(ctx context.Context, hashTopic string, topicID int) error
	Topic(ctx context.Context, hashTopic string) (string, error)
}

type SupportService struct {
	log *slog.Logger
	bot *telebot.Bot
	db DB
}

type Message struct {
	BotID int64
	ChatID int64
	UserID int64
	UserName string
	Payload string
}

func New(log *slog.Logger, db DB, token string, timeout int) *SupportService {
	bot, err := telebot.NewBot(
		telebot.Settings{
			Token: token,
			Poller: &telebot.LongPoller{Timeout: time.Second * time.Duration(timeout)},
		},
	)
	
	if err != nil {
		panic(err)
	}

	return &SupportService{
		log,
		bot,
		db,
	}
}

func(sbot *SupportService) ProcessMessage(msg string) {
	telegramMessage, err := parseMessage(msg)
	if err != nil {
		return
	}
	err = sbot.handleMessage(telegramMessage)
	if err != nil {
		sbot.manualContactingSupport(telegramMessage.ChatID)
	} else {
		sbot.bot.Send(telebot.ChatID(telegramMessage.ChatID), sendMessageToSupport)
	}
}

func(sbot *SupportService) handleMessage(telegramMessage Message) error {
	hash, err := hashTopic(telegramMessage.ChatID, telegramMessage.UserID)
	if err != nil {
		return err
	}

	topic, err := sbot.db.Topic(context.Background(), hash)
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

func (sbot *SupportService) manualContactingSupport(chatID int64) {
	inLine.Inline(
		inLine.Row(btnHelp),
	) 
	_, err := sbot.bot.Send(telebot.ChatID(chatID), cantCreateTopic, inLine)
	if err != nil {
		sbot.log.Error("SendInlineCommand", err)
	}
}

func hashTopic(chatID, userID int64) (string, error) {
	h := sha1.New()
    _, err := h.Write([]byte(fmt.Sprintf("%d", chatID + userID)))
	if err == nil {
    	sha1_hash := hex.EncodeToString(h.Sum(nil))
		return sha1_hash, nil
	}
	
	return "", err
}

func (sbot *SupportService) transferMessageToTopic(topic string, telegramMessage Message) error {
	id, err := strconv.ParseInt(topic, 10, 64)
	if err != nil {
		return err
	}
	group := telebot.ChatID(id)
	//if telegramMessage.Payload != "" {
	_, err = sbot.bot.Send(group, telegramMessage.Payload)
	return err	
	//}
}

func (sbot *SupportService) createTopic(telegramMessage Message) error {
	chat := &telebot.Chat{
		ID: telegramMessage.ChatID,
		Private: true,
	}

	topic, err := sbot.bot.CreateTopic(chat, generateTopic(telegramMessage.UserName))
	if err != nil {
		return err
	}

	hash, err := hashTopic(telegramMessage.ChatID, telegramMessage.UserID)
	if err != nil {
		return err
	}

	err = sbot.db.NewTopic(
		context.Background(),
		hash,
		topic.ThreadID,
	)

	return err
}

func generateTopic(userName string) *telebot.Topic {
	return &telebot.Topic{
			Name: fmt.Sprintf("Chat with %s", userName),
			IconColor: 0,
		}
}