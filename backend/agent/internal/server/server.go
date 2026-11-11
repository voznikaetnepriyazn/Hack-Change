package server

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/agent"
)

const completedStreamRetention = time.Minute

type Server struct {
	app      *fiber.App
	agent    *agent.Agent
	channels map[string]*activeStream
	mu       sync.Mutex
}

type activeStream struct {
	channel    chan string
	cancel     context.CancelFunc
	replaced   chan struct{}
	finished   chan struct{}
	subscribed bool
}

func NewServer(a *agent.Agent) *Server {
	app := fiber.New()

	s := &Server{
		app:      app,
		agent:    a,
		channels: make(map[string]*activeStream),
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // especially for hackaton
		AllowHeaders: "Origin, Content-Type, Accept, X-Session-ID",
		AllowMethods: "GET, POST, OPTIONS",
	}))

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
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
	s.app.Post("/chat", func(c *fiber.Ctx) error {
		var req struct {
			Message string `json:"message"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		sessionID := strings.Clone(c.Get("X-Session-ID", "default-user")) // frontend must send ID
		message := strings.Clone(req.Message)
		ctx, cancel := context.WithCancel(context.Background())
		stream := &activeStream{
			channel:  make(chan string, 100),
			cancel:   cancel,
			replaced: make(chan struct{}),
			finished: make(chan struct{}),
		}

		s.mu.Lock()
		previous := s.channels[sessionID]
		s.channels[sessionID] = stream
		s.mu.Unlock()

		if previous != nil {
			previous.cancel()
			close(previous.replaced)
		}

		go func() {
			if previous != nil {
				<-previous.finished
			}

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
	})

	// andpoint for SSE
	s.app.Get("/stream", func(c *fiber.Ctx) error {
		sessionID := strings.Clone(c.Query("sessionId", "default-user"))

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

				case <-stream.replaced:
					return
				}
			}
		})

		return nil
	})
}

func (s *Server) Start(port string) error {
	return s.app.Listen(":" + port)
}
