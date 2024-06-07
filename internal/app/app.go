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

func New(log *slog.Logger, dbHost, dbPort, dbPassword, token string, timeout int, chatID int64) App {
	db, err := redis.New(log, dbHost, dbPort, dbPassword)
	if err != nil {
		panic(err)
	}
	botService := supportline.New(log, db, token, timeout, chatID)
	appsupport := appsupport.New(log, db, botService)
	
	return App{appsupport}
}