package appsupportline

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"golang.org/x/net/websocket"

	supp "github.com/behummble/support_line_bot/internal/service/support_line"
)

const (
	userMessages = "/user/message"
	supportMessages = "/support/message"
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

func (support *Support) ListenMessages(host string, port int) {
	http.Handle(userMessages, websocket.Handler(support.userMessage))
	http.Handle(supportMessages, websocket.Handler(support.supportMessage))
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func (support *Support) RemoveTopics() {
	support.supportService.RemoveTopics()
}

func (support *Support) userMessage(ws *websocket.Conn) {
	var data []byte
	err := websocket.Message.Receive(ws, &data)
	if err == nil {
		support.supportService.ProcessUserMessage(string(data))
	} else {
		support.log.Error("HandleWebSocketMessage", err)
	}
}

func (support *Support) supportMessage(ws *websocket.Conn) {
	var data []byte
	err := websocket.Message.Receive(ws, &data)
	if err == nil {
		support.supportService.ProcessSupportMessage(data)
	} else {
		support.log.Error("HandleWebSocketMessage", err)
	}
}