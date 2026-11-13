package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/domain"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/service"
	authsession "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/session"
)

type testAuthenticator struct {
	user domain.User
	err  error
}

func (a testAuthenticator) Authenticate(string, string) (domain.User, error) {
	return a.user, a.err
}

type testSessions struct {
	values map[string]authsession.Session
	ttl    time.Duration
	fail   bool
}

func newTestSessions() *testSessions {
	return &testSessions{values: make(map[string]authsession.Session)}
}

func (s *testSessions) Create(_ context.Context, userID string, ttl time.Duration) (string, error) {
	if s.fail {
		return "", errors.New("session store unavailable")
	}
	s.ttl = ttl
	token := "opaque-test-token"
	s.values[token] = authsession.Session{UserID: userID}
	return token, nil
}

func (s *testSessions) Get(_ context.Context, id string) (authsession.Session, error) {
	session, ok := s.values[id]
	if !ok {
		return authsession.Session{}, authsession.ErrNotFound
	}
	return session, nil
}

func (s *testSessions) Delete(_ context.Context, id string) error {
	delete(s.values, id)
	return nil
}

func newTestHandler(t *testing.T, auth Authenticator, sessions authsession.Store) *Handler {
	t.Helper()
	h, err := New(auth, sessions, Config{
		CookieName: "auth_session",
		SessionTTL: 2 * time.Hour,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	return h
}

func TestNewRejectsInsecureSameSiteNoneCookie(t *testing.T) {
	h, err := New(
		testAuthenticator{},
		newTestSessions(),
		Config{
			CookieName: "auth_session",
			SessionTTL: time.Hour,
			SameSite:   http.SameSiteNoneMode,
		},
	)
	if h != nil || !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("New() = (%v, %v), want nil handler and invalid config error", h, err)
	}
}

func TestLoginCreatesSessionAndHttpOnlyCookie(t *testing.T) {
	sessions := newTestSessions()
	auth := testAuthenticator{user: domain.User{
		ID:       "user-1",
		FullName: "Alex Intern",
		Login:    "alex",
		Role:     domain.RoleIntern,
	}}
	h := newTestHandler(t, auth, sessions)
	app := fiber.New()
	h.RegisterRoutes(app)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(
		`{"login":"alex","password":"secret"}`,
	))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	cookies := resp.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("login set %d cookies, want 1", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != "auth_session" || cookie.Value == "" {
		t.Errorf("login cookie = %+v, want named non-empty token", cookie)
	}
	if !cookie.HttpOnly || cookie.Secure || cookie.Path != "/" || cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("login cookie security attributes = %+v", cookie)
	}
	if cookie.MaxAge != int((2 * time.Hour).Seconds()) {
		t.Errorf("cookie MaxAge = %d, want %d", cookie.MaxAge, int((2 * time.Hour).Seconds()))
	}
	if got := sessions.values[cookie.Value].UserID; got != "user-1" {
		t.Errorf("stored session user ID = %q, want user-1", got)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read login response body: %v", err)
	}
	if strings.Contains(string(body), "password") {
		t.Error("login response must not expose password fields")
	}
}

func TestLoginRejectsInvalidCredentialsWithoutCookie(t *testing.T) {
	h := newTestHandler(t, testAuthenticator{err: service.ErrInvalidCredentials}, newTestSessions())
	app := fiber.New()
	h.RegisterRoutes(app)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(
		`{"login":"alex","password":"wrong"}`,
	))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() returned error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("login status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if len(resp.Cookies()) != 0 {
		t.Errorf("failed login set cookies: %+v", resp.Cookies())
	}
}

func TestLoginAcceptsManagerRole(t *testing.T) {
	auth := testAuthenticator{user: domain.User{
		ID:       "manager-1",
		FullName: "Dana Manager",
		Login:    "dana",
		Role:     domain.RoleManager,
	}}
	h := newTestHandler(t, auth, newTestSessions())
	app := fiber.New()
	h.RegisterRoutes(app)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(
		`{"login":"dana","password":"secret"}`,
	))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() returned error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read login response body: %v", err)
	}
	if !strings.Contains(string(body), `"role":"manager"`) {
		t.Errorf("login response = %s, want manager role", body)
	}
}

func TestLogoutRevokesSessionAndExpiresCookie(t *testing.T) {
	sessions := newTestSessions()
	sessions.values["opaque-test-token"] = authsession.Session{UserID: "user-1"}
	h := newTestHandler(t, testAuthenticator{}, sessions)
	app := fiber.New()
	h.RegisterRoutes(app)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "auth_session", Value: "opaque-test-token"})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() returned error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("logout status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
	if _, ok := sessions.values["opaque-test-token"]; ok {
		t.Error("logout did not revoke the stored session")
	}
	cookies := resp.Cookies()
	if len(cookies) != 1 || cookies[0].Name != "auth_session" || !cookies[0].Expires.Before(time.Now()) {
		t.Errorf("logout cookies = %+v, want expired auth cookie", cookies)
	}
}
