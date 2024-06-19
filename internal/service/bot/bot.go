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
		log.Error("InitializeBot", err)
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