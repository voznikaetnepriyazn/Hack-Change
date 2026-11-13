package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/agent"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/domain"
	authhandler "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/handler"
	authmiddleware "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/middleware"
)

const completedStreamRetention = time.Minute

var ErrInvalidAuthConfig = errors.New("invalid authenticated server configuration")

// AuthConfig contains the authentication dependencies required by protected routes.
type AuthConfig struct {
	Handler        *authhandler.Handler
	Middleware     fiber.Handler
	AllowedOrigins []string
}

type Server struct {
	app      *fiber.App
	agent    *agent.Agent
	channels map[string]*activeStream
	mu       sync.Mutex
}

type activeStream struct {
	channel    chan string
	cancel     context.CancelFunc
	finished   chan struct{}
	subscribed bool
}

// NewServer creates a server without authentication. Use only for tests or local development.
// Production servers must be created with NewServerWithAuth.
func NewServer(a *agent.Agent) *Server {
	return newServer(a, nil, nil, false, "*", false)
}

// NewServerWithAuth creates a server whose chat routes require an authenticated intern.
func NewServerWithAuth(a *agent.Agent, cfg AuthConfig) (*Server, error) {
	if cfg.Handler == nil || cfg.Middleware == nil || len(cfg.AllowedOrigins) == 0 {
		return nil, ErrInvalidAuthConfig
	}
	for _, origin := range cfg.AllowedOrigins {
		if strings.TrimSpace(origin) == "" || strings.TrimSpace(origin) == "*" {
			return nil, ErrInvalidAuthConfig
		}
	}
	return newServer(
		a,
		cfg.Handler,
		cfg.Middleware,
		true,
		strings.Join(cfg.AllowedOrigins, ","),
		true,
	), nil
}

func newServer(a *agent.Agent, authHandler *authhandler.Handler, authMiddleware fiber.Handler, requireAuth bool, allowedOrigins string, allowCredentials bool) *Server {
	app := fiber.New()

	s := &Server{
		app:      app,
		agent:    a,
		channels: make(map[string]*activeStream),
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowMethods:     "GET, POST, OPTIONS",
		AllowCredentials: allowCredentials,
	}))

	s.setupRoutes(authHandler, authMiddleware, requireAuth)
	return s
}

func (s *Server) setupRoutes(authHandler *authhandler.Handler, authMiddleware fiber.Handler, requireAuth bool) {
	if authHandler != nil {
		authHandler.RegisterRoutes(s.app)
	}

	// basic heailth check endpoint
	s.app.Get("/health", func(c *fiber.Ctx) error {
		s.mu.Lock()
		activeStreams := len(s.channels)
		s.mu.Unlock()

		return c.JSON(fiber.Map{
			"status":         "running",
			"server_time":    time.Now().Format(time.RFC3339),
			"active_streams": activeStreams,
			"agent":          "ready",
			"version":        "1.0.0",
		})
	})

	// andpoint for message sending
	chatHandler := func(c *fiber.Ctx) error {
		var req struct {
			Message string `json:"message"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		sessionID := strings.Clone(c.Get("X-Session-ID"))
		if requireAuth {
			sessionID, _ = authmiddleware.UserID(c)
		}
		if sessionID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Session ID is required"})
		}
		message := strings.Clone(req.Message)
		s.mu.Lock()
		if _, exists := s.channels[sessionID]; exists {
			s.mu.Unlock()
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "A response for this session is already in progress or awaiting delivery",
			})
		}
		ctx, cancel := context.WithCancel(context.Background())
		stream := &activeStream{
			channel:  make(chan string, 100),
			cancel:   cancel,
			finished: make(chan struct{}),
		}
		s.channels[sessionID] = stream
		s.mu.Unlock()

		go func() {
			defer func() {
				close(stream.channel)
				cancel()
				s.mu.Lock()
				if s.channels[sessionID] == stream {
					if stream.subscribed {
						delete(s.channels, sessionID)
					} else {
						time.AfterFunc(completedStreamRetention, func() {
							s.mu.Lock()
							if s.channels[sessionID] == stream && !stream.subscribed {
								delete(s.channels, sessionID)
							}
							s.mu.Unlock()
						})
					}
				}
				s.mu.Unlock()
				close(stream.finished)
			}()

			_, err := s.agent.RunWithChannel(ctx, sessionID, message, stream.channel)
			if err != nil {
				slog.Error("Agent failed", "error", err)
			}
		}()

		return c.SendStatus(fiber.StatusAccepted)
	}
	if requireAuth {
		s.app.Post("/chat", authMiddleware, requireIntern, chatHandler)
	} else {
		s.app.Post("/chat", chatHandler)
	}

	// andpoint for SSE
	streamHandler := func(c *fiber.Ctx) error {
		sessionID := strings.Clone(c.Query("sessionId"))
		if requireAuth {
			sessionID, _ = authmiddleware.UserID(c)
		}
		if sessionID == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Session ID is required")
		}

		s.mu.Lock()
		stream := s.channels[sessionID]
		if stream != nil && stream.subscribed {
			s.mu.Unlock()
			return c.Status(fiber.StatusConflict).SendString("Stream already has a subscriber")
		}
		if stream != nil {
			stream.subscribed = true
		}
		s.mu.Unlock()

		if stream == nil {
			return c.Status(404).SendString("Session not found")
		}

		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("X-Accel-Buffering", "no")

		c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			completed := false
			defer func() {
				if !completed {
					s.mu.Lock()
					if s.channels[sessionID] == stream {
						stream.subscribed = false
					}
					s.mu.Unlock()
				}
			}()

			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case text, ok := <-stream.channel:
					if !ok {
						fmt.Fprintf(w, "event: done\ndata: [DONE]\n\n")
						if err := w.Flush(); err != nil {
							return
						}
						completed = true
						s.mu.Lock()
						if s.channels[sessionID] == stream {
							delete(s.channels, sessionID)
						}
						s.mu.Unlock()
						return
					}

					s.mu.Lock()
					if s.channels[sessionID] != stream {
						s.mu.Unlock()
						return
					}

					s.mu.Unlock()
					text = strings.ReplaceAll(text, "\r\n", "\n")
					text = strings.ReplaceAll(text, "\r", "\n")
					for _, line := range strings.Split(text, "\n") {
						if _, err := fmt.Fprintf(w, "data: %s\n", line); err != nil {
							return
						}
					}
					if _, err := fmt.Fprint(w, "\n"); err != nil {
						return
					}
					if err := w.Flush(); err != nil {
						return
					}

				case <-ticker.C:
					if _, err := fmt.Fprintf(w, ": keep-alive\n\n"); err != nil {
						return
					}
					if err := w.Flush(); err != nil {
						return
					}
				}
			}
		})

		return nil
	}
	if requireAuth {
		s.app.Get("/stream", authMiddleware, requireIntern, streamHandler)
	} else {
		s.app.Get("/stream", streamHandler)
	}
}

func requireIntern(c *fiber.Ctx) error {
	role, ok := authmiddleware.UserRole(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authentication required"})
	}
	if role != domain.RoleIntern {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Chat access is restricted"})
	}
	return c.Next()
}

func (s *Server) Start(port string) error {
	return s.app.Listen(":" + port)
}
