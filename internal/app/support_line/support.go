package appsupportline

import (
	"log/slog"

	supp "github.com/behummble/support_line_bot/internal/service/support_line"
	
)

type Router interface {
	Serve(host string, port int)
	Register()
}

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

func (support *Support) Register() {
	support.router.Register()
}

func (support *Support) ListenMessages(host string, port int) {
	support.log.Info("Start listening messages")
	support.router.Serve(host, port)
}

func (support *Support) RemoveTopics() {
	support.supportService.RemoveTopics()
}
