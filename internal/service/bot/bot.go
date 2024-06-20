package bot

import (
	"log/slog"
	"time"
	"gopkg.in/telebot.v3"
	"github.com/behummble/support_line_bot/pkg/crypto"
)

type Bot struct {
	log *slog.Logger
	token string
	client *telebot.Bot
}

func New(log *slog.Logger, encryptedToken []byte, timeout int) (*Bot, error) {
	token, err := crypto.DecryptData(encryptedToken)
	if err != nil {
		return nil, err
	}

	client, err := newBotClient(token, timeout)
	if err != nil {
		return nil, err
	}

	return &Bot{
		log: log,
		token: token,
		client: client,
	}, nil
}

func newBotClient(token string, timeout int) (*telebot.Bot, error) {
	bot, err := telebot.NewBot(
		telebot.Settings{
			Token: token,
			Poller: &telebot.LongPoller{Timeout: time.Second * time.Duration(timeout)},
		},
	)
	return bot, err
}

func (bot *Bot) Token() string {
	return bot.token
}

func (bot *Bot) ChatByID(chatID int64) (*telebot.Chat, error) {
	return bot.client.ChatByID(chatID)
}

func (bot *Bot) Forward(to telebot.Recipient, msg telebot.Editable, opts ...interface{}) (*telebot.Message, error) {
	return bot.client.Forward(to, msg, opts)
}

func (bot *Bot) Send(to telebot.Recipient, what interface{}, opts ...interface{}) (*telebot.Message, error) {
	return bot.client.Send(to, what, opts)
}

func (bot *Bot) CreateTopic(chat *telebot.Chat, topic *telebot.Topic) (*telebot.Topic, error) {
	return bot.client.CreateTopic(chat, topic)
}

func (bot *Bot) CloseTopic(chat *telebot.Chat, topic *telebot.Topic) error {
	return bot.client.CloseTopic(chat, topic)
}