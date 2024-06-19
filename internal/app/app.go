package app

import (
	"log/slog"

	appsupport "github.com/behummble/support_line_bot/internal/app/support_line"
	"github.com/behummble/support_line_bot/internal/repo/db/redis"
	"github.com/behummble/support_line_bot/internal/service/support_line"
	"github.com/behummble/support_line_bot/internal/websocket/updates"
	"github.com/behummble/support_line_bot/internal/config"
)

type App struct {
	Bot *appsupport.Support
}

func New(log *slog.Logger, config *config.Config) App {
	db, err := redis.New(
		log, 
		config.Redis.Host, 
		config.Redis.Port, 
		config.Redis.Password)

	if err != nil {
		panic(err)
	}
	
	botService := supportline.New(log, db, config.Bot.ChatID, config.Bot.UpdateTimeout)
	router := updates.New(log, botService)
	appsupport := appsupport.New(log, botService, router)
	
	return App{Bot: appsupport}
}