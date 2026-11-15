package middleware

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/domain"
	authsession "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/session"
)

const (
	userIDLocalKey = "auth.userID"
	roleLocalKey   = "auth.role"
)

var ErrInvalidConfig = errors.New("invalid authentication middleware configuration")

// Config defines how the middleware reads the authentication cookie.
type Config struct {
	CookieName string
}

// UserResolver resolves the current user record by its stable ID.
type UserResolver interface {
	FindByID(id string) (domain.User, bool)
}

// New creates middleware that resolves the authentication cookie through the session store.
func New(store authsession.Store, users UserResolver, cfg Config) (fiber.Handler, error) {
	if store == nil || users == nil || strings.TrimSpace(cfg.CookieName) == "" {
		return nil, ErrInvalidConfig
	}

	return func(c *fiber.Ctx) error {
		token := c.Cookies(cfg.CookieName)
		if token == "" {
			return unauthorized(c)
		}

		authSession, err := store.Get(c.UserContext(), token)
		if errors.Is(err, authsession.ErrNotFound) {
			return unauthorized(c)
		}
		if err != nil {
			slog.Error("Failed to resolve authentication session", "error", err)
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Authentication service unavailable",
			})
		}

		user, ok := users.FindByID(authSession.UserID)
		if !ok || user.ID == "" {
			if err := store.Delete(c.UserContext(), token); err != nil {
				slog.Warn("Failed to revoke session for missing user", "error", err)
			}
			return unauthorized(c)
		}

		c.Locals(userIDLocalKey, user.ID)
		c.Locals(roleLocalKey, user.Role)
		return c.Next()
	}, nil
}

// UserID returns the authenticated user ID set by this middleware.
func UserID(c *fiber.Ctx) (string, bool) {
	userID, ok := c.Locals(userIDLocalKey).(string)
	return userID, ok && userID != ""
}

// UserRole returns the authenticated user's current role.
func UserRole(c *fiber.Ctx) (domain.Role, bool) {
	role, ok := c.Locals(roleLocalKey).(domain.Role)
	return role, ok && role != ""
}

func unauthorized(c *fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": "Authentication required",
	})
}
