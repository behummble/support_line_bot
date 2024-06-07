package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/behummble/support_line_bot/internal/config"
	"github.com/behummble/support_line_bot/internal/app"
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
	)
	app.Bot.StartListenUpdates(config.Bot.Name)	
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