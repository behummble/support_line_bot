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
		config,
	)
	go app.Bot.RemoveTopics()
	app.Bot.Register()
	app.Bot.ListenMessages(config.Server.Host, config.Server.Port)
}

func initLog() *slog.Logger {
	log := slog.New(slog.NewJSONHandler(
		os.Stdout, 
		&slog.HandlerOptions{Level: slog.LevelDebug}))

	return log
}

func setEnv() {
	err := godotenv.Load("app.env")
	if err != nil {
		panic(err)
	}
}