package supportline

import (
	"context"
	"time"
	"log/slog"

	"gopkg.in/telebot.v3"
)

const (
	errorMessage = "Sorry, something went wrong, try to repeat the request later"
	startMessage = "Hi, %s"
)

const (
	commandNotFound = "Can`t find command %s"
)

type MessageSaver interface {
	Save(ctx context.Context, botName, msg string) error
}

type SupportService struct {
	log *slog.Logger
	bot *telebot.Bot
	Saver MessageSaver
}

type Message struct {
	BotID int64
	ChatID int64
	UserID int64
	UserName string
	Payload string
}

func New(log *slog.Logger, saver MessageSaver, token string, timeout int) *SupportService {
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
		saver,
	}
}

func(sbot *SupportService) ProcessMessage(msg string) {
	err := sbot.handleMessage(msg)
	if err != nil {
		sbot.manualContactingSupport()
	}
}

func(sbot *SupportService) handleMessage(upd string) (error) {
	/* switch event and return 
	if upd.Message.IsCommand() {
		return handleCommand(upd)
	}
	
	if upd.Message.Text != "" {
		msg, err := prepareMessage(upd)
		if err == nil {
			err = sbot.Saver.Save(
				context.Background(),
				botName,
				msg,
			)
		}
		
		return nil
	} */

	return nil
}

func handleCommand(upd string) (error) {
	return nil
}

func prepareMessage(upd string) (error) {
	/*msg := Message {
		upd.Message.ViaBot.ID,
		upd.Message.Chat.ID,
		upd.Message.From.ID,
		upd.Message.From.UserName,
		upd.Message.Text,
	}

	payload, err := json.Marshal(msg)
	return string(payload), err */
	return nil
}
/*
func handleStart(upd tgbotapi.Update) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(
		upd.Message.Chat.ID, 
		fmt.Sprintf(startMessage, upd.Message.From.UserName),
	)
}

func handleAction1(upd tgbotapi.Update) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(
		upd.Message.Chat.ID, 
		"Thanks for your help",
	)
}

func handleAction2(upd tgbotapi.Update) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(
		upd.Message.Chat.ID, 
		"Your phone belongs to me now",
	)
}

func errorResponse(chatID int64) (tgbotapi.MessageConfig) {
	return tgbotapi.NewMessage(chatID, errorMessage)
}
*/

func (sbot *SupportService) manualContactingSupport() {

}