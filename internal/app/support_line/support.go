package appsupportline

import (
	"context"
	"log/slog"

	supp "github.com/behummble/support_line_bot/internal/service/support_line"
)

type MessageReceiver interface {
	Receive(ctx context.Context, botName string, msgs chan<- string)
}

type Support struct {
	log *slog.Logger
	receiver MessageReceiver
	supportService *supp.SupportService
	
}

func New(log *slog.Logger, receiver MessageReceiver, botService *supp.SupportService) (*Support) {
	return &Support{log, receiver, botService}
} 

func(support *Support) StartListenUpdates(botName string) {
	msgs := make(chan string)
	go support.receiver.Receive(context.Background(), botName, msgs)
	for {
		msg := <- msgs
		support.supportService.ProcessMessage(msg)
	}
}

func (support *Support) RemoveTopics() {
	support.supportService.RemoveTopics()
}