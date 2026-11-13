package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/domain"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/service"
	authsession "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/session"
)

var ErrInvalidConfig = errors.New("invalid authentication handler configuration")

// Authenticator verifies user credentials.
type Authenticator interface {
	Authenticate(login, password string) (domain.User, error)
}

// Config controls the authentication cookie and session lifetime.
type Config struct {
	CookieName string
	SessionTTL time.Duration
	Secure     bool
	SameSite   http.SameSite
}

// Handler implements login and logout endpoints.
type Handler struct {
	auth     Authenticator
	sessions authsession.Store
	config   Config
}

// New creates an authentication HTTP handler.
func New(auth Authenticator, sessions authsession.Store, cfg Config) (*Handler, error) {
	if auth == nil || sessions == nil || strings.TrimSpace(cfg.CookieName) == "" || cfg.SessionTTL <= 0 {
		return nil, ErrInvalidConfig
	}
	if cfg.SameSite == 0 {
		cfg.SameSite = http.SameSiteLaxMode
	}
	if cfg.SameSite != http.SameSiteLaxMode &&
		cfg.SameSite != http.SameSiteStrictMode &&
		cfg.SameSite != http.SameSiteNoneMode {
		return nil, ErrInvalidConfig
	}
	if cfg.SameSite == http.SameSiteNoneMode && !cfg.Secure {
		return nil, ErrInvalidConfig
	}
	return &Handler{auth: auth, sessions: sessions, config: cfg}, nil
}

// RegisterRoutes registers public login and logout routes.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	router.Post("/auth/login", h.login)
	router.Post("/auth/logout", h.logout)
}

func (h *Handler) login(c *fiber.Ctx) error {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}
	if strings.TrimSpace(req.Login) == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Login and password are required"})
	}

	user, err := h.auth.Authenticate(req.Login, req.Password)
	if errors.Is(err, service.ErrInvalidCredentials) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid login or password"})
	}
	if err != nil {
		slog.Error("Authentication failed", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Authentication failed"})
	}
	if user.ID == "" || (user.Role != domain.RoleIntern && user.Role != domain.RoleLead && user.Role != domain.RoleManager) {
		slog.Error("Authenticator returned invalid user identity", "user_id", user.ID, "role", user.Role)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Authentication failed"})
	}

	token, err := h.sessions.Create(c.UserContext(), user.ID, h.config.SessionTTL)
	if err != nil {
		slog.Error("Failed to create authentication session", "error", err)
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Authentication service unavailable",
		})
	}

	c.Cookie(h.cookie(token))
	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"id":        user.ID,
			"full_name": user.FullName,
			"login":     user.Login,
			"role":      user.Role,
		},
	})
}

func (h *Handler) logout(c *fiber.Ctx) error {
	token := c.Cookies(h.config.CookieName)
	if token != "" {
		if err := h.sessions.Delete(c.UserContext(), token); err != nil {
			slog.Error("Failed to revoke authentication session", "error", err)
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Authentication service unavailable",
			})
		}
	}
	c.Cookie(h.expiredCookie())
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) cookie(token string) *fiber.Cookie {
	return &fiber.Cookie{
		Name:     h.config.CookieName,
		Value:    token,
		Path:     "/",
		HTTPOnly: true,
		Secure:   h.config.Secure,
		SameSite: sameSiteValue(h.config.SameSite),
		MaxAge:   int(h.config.SessionTTL.Seconds()),
		Expires:  time.Now().Add(h.config.SessionTTL),
	}
}

func (h *Handler) expiredCookie() *fiber.Cookie {
	return &fiber.Cookie{
		Name:     h.config.CookieName,
		Value:    "",
		Path:     "/",
		HTTPOnly: true,
		Secure:   h.config.Secure,
		SameSite: sameSiteValue(h.config.SameSite),
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
	}
}

func sameSiteValue(mode http.SameSite) string {
	switch mode {
	case http.SameSiteStrictMode:
		return "Strict"
	case http.SameSiteNoneMode:
		return "None"
	default:
		return "Lax"
	}
}
