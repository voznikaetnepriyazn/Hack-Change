package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/domain"
	authsession "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/session"
)

type failingStore struct{}

func (failingStore) Create(context.Context, string, time.Duration) (string, error) {
	return "", errors.New("store unavailable")
}

func (failingStore) Get(context.Context, string) (authsession.Session, error) {
	return authsession.Session{}, errors.New("store unavailable")
}

func (failingStore) Delete(context.Context, string) error {
	return errors.New("store unavailable")
}

type testUsers map[string]domain.User

func (u testUsers) FindByID(id string) (domain.User, bool) {
	user, ok := u[id]
	return user, ok
}

func internResolver() testUsers {
	return testUsers{
		"user-42": {
			ID:   "user-42",
			Role: domain.RoleIntern,
		},
	}
}

func newTestStore(t *testing.T) (*authsession.RedisStore, *redis.Client) {
	t.Helper()

	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close Redis client: %v", err)
		}
	})
	store, err := authsession.NewRedisStore(client, "")
	if err != nil {
		t.Fatalf("NewRedisStore() returned error: %v", err)
	}
	return store, client
}

func TestNewRequiresStoreAndCookieName(t *testing.T) {
	if handler, err := New(nil, internResolver(), Config{CookieName: "auth"}); err == nil || handler != nil {
		t.Errorf("New(nil, ...) = (%v, %v), want nil handler and error", handler, err)
	}

	store, _ := newTestStore(t)
	if handler, err := New(store, internResolver(), Config{}); err == nil || handler != nil {
		t.Errorf("New(..., empty cookie) = (%v, %v), want nil handler and error", handler, err)
	}
}

func TestMiddlewareRejectsMissingOrUnknownCookie(t *testing.T) {
	store, _ := newTestStore(t)
	handler, err := New(store, internResolver(), Config{CookieName: "auth_session"})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	app := fiber.New()
	app.Get("/private", handler, func(c *fiber.Ctx) error {
		t.Error("protected handler must not run without a valid session")
		return c.SendStatus(fiber.StatusOK)
	})

	for _, tt := range []struct {
		name   string
		cookie string
	}{
		{name: "missing cookie"},
		{name: "unknown token", cookie: "invalid-token"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/private", nil)
			if tt.cookie != "" {
				req.AddCookie(&http.Cookie{Name: "auth_session", Value: tt.cookie})
			}
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test() returned error: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != fiber.StatusUnauthorized {
				t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
			}
		})
	}
}

func TestMiddlewareResolvesSessionAndSetsUserID(t *testing.T) {
	store, _ := newTestStore(t)
	token, err := store.Create(context.Background(), "user-42", time.Hour)
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}
	handler, err := New(store, internResolver(), Config{CookieName: "auth_session"})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	app := fiber.New()
	app.Get("/private", handler, func(c *fiber.Ctx) error {
		userID, ok := UserID(c)
		if !ok {
			t.Error("UserID() did not return authenticated user")
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		if userID != "user-42" {
			t.Errorf("UserID() = %q, want user-42", userID)
		}
		if role, ok := UserRole(c); !ok || role != domain.RoleIntern {
			t.Errorf("UserRole() = (%q, %t), want intern", role, ok)
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.AddCookie(&http.Cookie{Name: "auth_session", Value: token})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() returned error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusNoContent {
		t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusNoContent)
	}
}

func TestMiddlewareReturnsUnavailableWhenStoreFails(t *testing.T) {
	handler, err := New(failingStore{}, internResolver(), Config{CookieName: "auth_session"})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	app := fiber.New()
	app.Get("/private", handler, func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.AddCookie(&http.Cookie{Name: "auth_session", Value: "some-token"})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() returned error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusServiceUnavailable)
	}
}
