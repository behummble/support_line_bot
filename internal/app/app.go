package app

import (
	"log/slog"

	appsupport "github.com/behummble/support_line_bot/internal/app/support_line"
	"github.com/behummble/support_line_bot/internal/repo/db/redis"
	"github.com/behummble/support_line_bot/internal/service/support_line"
)

type App struct {
	Bot *appsupport.Support
}

func New(log *slog.Logger, token, dbHost, dbPort, dbPassword string, timeout int, chatID int64) App {
	db, err := redis.New(log, dbHost, dbPort, dbPassword)
	if err != nil {
		panic(err)
	}
	bot, err := telebot.NewBot(
		telebot.Settings{
			Token: token,
			Poller: &telebot.LongPoller{Timeout: time.Second * time.Duration(timeout)},
		},
	)
	if err != nil {
		panic(err)
	}

	botService := supportline.New(log, db, chatID)
	appsupport := appsupport.New(log, db, botService)
	
	return App{appsupport}
}