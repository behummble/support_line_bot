package updates

import(
	"net/http"
	"log/slog"
	"fmt"
	"golang.org/x/net/websocket"
	"github.com/behummble/support_line_bot/internal/service/support_line"
)

const (
	userMessages = "/user/message"
	supportMessages = "/support/message"
	ping = "/ping"
)

type Router struct {
	supportService *supportline.Support
	log *slog.Logger
	mux *http.ServeMux
}

func New(log *slog.Logger, support *supportline.Support) *Router {
	m := http.NewServeMux()
	return &Router{
		supportService: support,
		log: log,
		mux: m,
	}
}

func (r *Router) Register() {
	r.mux.Handle(userMessages, websocket.Handler(r.userMessage))
	r.mux.Handle(supportMessages, websocket.Handler(r.supportMessage))
	r.mux.Handle(ping, websocket.Handler(
		func(ws *websocket.Conn) {
			websocket.Message.Send(ws, "pong")
		},
	))
}

func (r *Router) Serve(host string, port int) {
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), r.mux)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func (r *Router) userMessage(ws *websocket.Conn) {
	var data []byte
	err := websocket.Message.Receive(ws, &data)
	if err == nil {
		r.supportService.ProcessUserMessage(data)
	} else {
		r.log.Error("HandleWebSocketMessage", err)
	}
}

func (r *Router) supportMessage(ws *websocket.Conn) {
	var data []byte
	err := websocket.Message.Receive(ws, &data)
	if err == nil {
		r.supportService.ProcessSupportMessage(data)
	} else {
		r.log.Error("HandleWebSocketMessage", err)
	}
}