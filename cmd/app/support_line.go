package main

import (
	"log/slog"
	"os"

	"github.com/behummble/support_line_bot/internal/app"
	"github.com/behummble/support_line_bot/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	log := initLog()
	setEnv()
	config := config.MustLoad()
	app := app.New(
		log, 
		config.Bot.Token, 
		config.Redis.Host, 
		config.Redis.Port,
		config.Redis.Password,
		config.Bot.UpdateTimeout,
		config.Bot.ChatID,
	)
	go app.Bot.RemoveTopics()
	go app.Bot.ListenUpdates(config.Bot.Name)
	app.Bot.ListenSupportMessages(
		config.Server.Host, 
		config.Server.Port,
		config.Server.Path)
}

func initLog() *slog.Logger {
	log := slog.New(slog.NewJSONHandler(
		os.Stdout, 
		&slog.HandlerOptions{Level: slog.LevelDebug}))

	return log
}

func setEnv() {
	err := godotenv.Load("../../.env")
	if err != nil {
		panic(err)
	}
}