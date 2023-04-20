package routes

import (
	"github.com/abergasov/market_timer/internal/logger"
	"github.com/abergasov/market_timer/internal/service/notifier"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
	"go.uber.org/zap"
)

type Server struct {
	appAddr    string
	log        logger.AppLogger
	service    *notifier.Service
	httpEngine *fiber.App
}

// InitAppRouter initializes the HTTP Server.
func InitAppRouter(log logger.AppLogger, service *notifier.Service, address string) *Server {
	app := &Server{
		appAddr:    address,
		httpEngine: fiber.New(fiber.Config{}),
		service:    service,
		log:        log.With(zap.String("service", "http")),
	}
	app.httpEngine.Use(recover.New())
	app.initRoutes()
	return app
}

func (s *Server) initRoutes() {
	s.httpEngine.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("pong")
	})
	s.httpEngine.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	s.httpEngine.Get("/ws/:chain/:percentage", websocket.New(s.handleTransactions))
}

// Run starts the HTTP Server.
func (s *Server) Run() error {
	s.log.Info("Starting HTTP server", zap.String("port", s.appAddr))
	return s.httpEngine.Listen(s.appAddr)
}

func (s *Server) Stop() {
	s.log.Info("stopping service")
	if err := s.httpEngine.Shutdown(); err != nil {
		s.log.Error("unable to stop http service", err)
	}
}
