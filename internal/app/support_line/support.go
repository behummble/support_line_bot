package appsupportline

import (
	"log/slog"

	supp "github.com/behummble/support_line_bot/internal/service/support_line"
	
)

const (
	userMessages = "/user/message"
	supportMessages = "/support/message"
)

type Router interface {
	Serve(host string, port int)
	Register()
}

/*type SupportService interface {
	ProcessUserMessage(message []byte)
	ProcessSupportMessage(message []byte)
} */

type Support struct {
	log *slog.Logger
	supportService *supp.Support
	router Router
}

func New(log *slog.Logger, support *supp.Support, router Router) (*Support) {
	return &Support{
		log: log,
		supportService: support,
		router: router,
	}
} 

func (support *Support) ListenMessages(host string, port int) {
	support.router.Serve(host, port)
}

func (support *Support) RemoveTopics() {
	support.supportService.RemoveTopics()
}
