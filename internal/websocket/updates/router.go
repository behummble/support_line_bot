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
)

type Router struct {
	supportService *supportline.Support
	log *slog.Logger
	router *http.ServeMux
}

func New(log *slog.Logger, support *supportline.Support) *Router {
	m := http.NewServeMux()
	return &Router{
		supportService: support,
		log: log,
		router: m,
	}
}

func (r *Router) Register() {
	r.router.Handle(userMessages, websocket.Handler(r.userMessage))
	r.router.Handle(supportMessages, websocket.Handler(r.supportMessage))
}

func (r *Router) Serve(host string, port int) {
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), r.router)
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