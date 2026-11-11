package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	openai "github.com/sashabaranov/go-openai"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/agent"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/domain"
	authhandler "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/handler"
	authmiddleware "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/middleware"
	authsession "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/session"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/config"
	tool "github.com/voznikaetnepriyazn/Hack-Change/agent/tools"
)

type testAuthSessions map[string]authsession.Session

func (s testAuthSessions) Create(_ context.Context, userID string, _ time.Duration) (string, error) {
	token := "token-" + userID
	s[token] = authsession.Session{UserID: userID}
	return token, nil
}

func (s testAuthSessions) Get(_ context.Context, token string) (authsession.Session, error) {
	session, ok := s[token]
	if !ok {
		return authsession.Session{}, authsession.ErrNotFound
	}
	return session, nil
}

func (s testAuthSessions) Delete(_ context.Context, token string) error {
	delete(s, token)
	return nil
}

type testAuthUsers map[string]domain.User

func (u testAuthUsers) FindByID(id string) (domain.User, bool) {
	user, ok := u[id]
	return user, ok
}

type testAuthenticator struct{}

func (testAuthenticator) Authenticate(string, string) (domain.User, error) {
	return domain.User{}, errors.New("not used by this test")
}

func newTestServer(t *testing.T, llmHandler http.HandlerFunc) *Server {
	t.Helper()

	llmServer := httptest.NewServer(llmHandler)
	t.Cleanup(llmServer.Close)

	clientConfig := openai.DefaultConfig("test-token")
	clientConfig.BaseURL = llmServer.URL + "/v1"
	assistant := agent.NewAgent(
		openai.NewClientWithConfig(clientConfig),
		tool.NewRegistry(),
		"gpt-4o-mini",
		true,
	).WithLimits(config.Limits{
		MaxIterations: 2,
		Deadline:      2 * time.Second,
		LLMTimeout:    time.Second,
	})
	return NewServer(assistant)
}

func TestAuthenticatedServerRequiresInternAndIgnoresClientSessionIDs(t *testing.T) {
	sessions := testAuthSessions{
		"lead-token": authsession.Session{UserID: "lead-1"},
	}
	users := testAuthUsers{
		"lead-1": {
			ID:   "lead-1",
			Role: domain.RoleLead,
		},
	}
	authHandler, err := authhandler.New(testAuthenticator{}, sessions, authhandler.Config{
		CookieName: "auth_session",
		SessionTTL: time.Hour,
	})
	if err != nil {
		t.Fatalf("authhandler.New() returned error: %v", err)
	}
	authMiddleware, err := authmiddleware.New(sessions, users, authmiddleware.Config{
		CookieName: "auth_session",
	})
	if err != nil {
		t.Fatalf("authmiddleware.New() returned error: %v", err)
	}

	assistant := agent.NewAgent(nil, tool.NewRegistry(), "unused", false)
	s, err := NewServerWithAuth(assistant, AuthConfig{
		Handler:        authHandler,
		Middleware:     authMiddleware,
		AllowedOrigins: []string{"http://localhost:3000"},
	})
	if err != nil {
		t.Fatalf("NewServerWithAuth() returned error: %v", err)
	}

	unauthenticated := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"message":"hi"}`))
	unauthenticated.Header.Set("Content-Type", "application/json")
	unauthenticated.Header.Set("X-Session-ID", "attacker-controlled")
	unauthenticatedResp := fiberRequest(t, s.app, unauthenticated)
	unauthenticatedResp.Body.Close()
	if unauthenticatedResp.StatusCode != http.StatusUnauthorized {
		t.Errorf("unauthenticated /chat status = %d, want %d", unauthenticatedResp.StatusCode, http.StatusUnauthorized)
	}

	leadRequest := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"message":"hi"}`))
	leadRequest.Header.Set("Content-Type", "application/json")
	leadRequest.Header.Set("X-Session-ID", "attacker-controlled")
	leadRequest.AddCookie(&http.Cookie{Name: "auth_session", Value: "lead-token"})
	leadResp := fiberRequest(t, s.app, leadRequest)
	leadResp.Body.Close()
	if leadResp.StatusCode != http.StatusForbidden {
		t.Errorf("lead /chat status = %d, want %d", leadResp.StatusCode, http.StatusForbidden)
	}
}

func writeLLMStream(w http.ResponseWriter, content string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(http.StatusOK)
	chunk, err := json.Marshal(map[string]any{
		"choices": []any{
			map[string]any{
				"index": 0,
				"delta": map[string]string{"content": content},
			},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", chunk)
	fmt.Fprint(w, "data: {\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n")
	fmt.Fprint(w, "data: [DONE]\n\n")
}

func fiberRequest(t *testing.T, app *fiber.App, req *http.Request) *http.Response {
	t.Helper()
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("app.Test() returned error: %v", err)
	}
	return resp
}

func serveFiberApp(t *testing.T, app *fiber.App) (*http.Client, string) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for Fiber test server: %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = app.Listener(listener)
	}()
	t.Cleanup(func() {
		_ = app.ShutdownWithTimeout(time.Second)
		_ = listener.Close()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("Fiber test server did not stop")
		}
	})

	return &http.Client{Timeout: 5 * time.Second}, "http://" + listener.Addr().String()
}

func postChatHTTP(t *testing.T, client *http.Client, baseURL, sessionID, message string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(
		http.MethodPost,
		baseURL+"/chat",
		strings.NewReader(fmt.Sprintf(`{"message":%q}`, message)),
	)
	if err != nil {
		t.Fatalf("create /chat request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-ID", sessionID)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST /chat: %v", err)
	}
	return resp
}

func getStreamHTTP(t *testing.T, client *http.Client, baseURL, sessionID string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, baseURL+"/stream?sessionId="+sessionID, nil)
	if err != nil {
		t.Fatalf("create /stream request: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("GET /stream: %v", err)
	}
	return resp
}

func TestHealthReportsStatusAndActiveStreamCount(t *testing.T) {
	s := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeLLMStream(w, "hello")
	})

	resp := fiberRequest(t, s.app, httptest.NewRequest(http.MethodGet, "/health", nil))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/health status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body struct {
		Status        string `json:"status"`
		ActiveStreams int    `json:"active_streams"`
		Agent         string `json:"agent"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if body.Status != "running" || body.Agent != "ready" || body.ActiveStreams != 0 {
		t.Errorf("health response = %+v, want running/ready with no active streams", body)
	}
}

func TestChatRejectsMalformedRequest(t *testing.T) {
	s := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("LLM must not be called for malformed chat request")
	})

	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"message":`))
	req.Header.Set("Content-Type", "application/json")
	resp := fiberRequest(t, s.app, req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("/chat status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestStreamReturnsNotFoundForUnknownSession(t *testing.T) {
	s := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("LLM must not be called for an unknown session stream")
	})

	resp := fiberRequest(t, s.app, httptest.NewRequest(http.MethodGet, "/stream?sessionId=missing", nil))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("/stream status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestChatAndStreamDeliverSSEEvents(t *testing.T) {
	s := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeLLMStream(w, "hello")
	})
	client, baseURL := serveFiberApp(t, s.app)
	postResp := postChatHTTP(t, client, baseURL, "session-1", "say hello")
	postResp.Body.Close()
	if postResp.StatusCode != http.StatusAccepted {
		t.Fatalf("/chat status = %d, want %d", postResp.StatusCode, http.StatusAccepted)
	}

	s.mu.Lock()
	completedStream := s.channels["session-1"]
	s.mu.Unlock()
	if completedStream == nil {
		t.Fatal("completed stream was removed before the client could subscribe")
	}
	select {
	case <-completedStream.finished:
	case <-time.After(time.Second):
		t.Fatal("LLM stream did not finish")
	}

	streamResp := getStreamHTTP(t, client, baseURL, "session-1")
	defer streamResp.Body.Close()
	if streamResp.StatusCode != http.StatusOK {
		t.Fatalf("/stream status = %d, want %d", streamResp.StatusCode, http.StatusOK)
	}
	if got := streamResp.Header.Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
		t.Errorf("Content-Type = %q, want text/event-stream", got)
	}
	if streamResp.Header.Get("Cache-Control") != "no-cache" {
		t.Errorf("Cache-Control = %q, want no-cache", streamResp.Header.Get("Cache-Control"))
	}

	body, err := io.ReadAll(streamResp.Body)
	if err != nil {
		t.Fatalf("read SSE response: %v", err)
	}
	if got, want := string(body), "data: hello\n\nevent: done\ndata: [DONE]\n\n"; got != want {
		t.Errorf("SSE body = %q, want %q", got, want)
	}
}

func TestStreamEncodesNewlinesAsSSEDataLines(t *testing.T) {
	s := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeLLMStream(w, "line one\nline two")
	})
	client, baseURL := serveFiberApp(t, s.app)
	postResp := postChatHTTP(t, client, baseURL, "sse-lines", "say two lines")
	postResp.Body.Close()

	s.mu.Lock()
	completedStream := s.channels["sse-lines"]
	s.mu.Unlock()
	if completedStream == nil {
		t.Fatal("completed stream was removed before the client could subscribe")
	}
	select {
	case <-completedStream.finished:
	case <-time.After(time.Second):
		t.Fatal("LLM stream did not finish")
	}

	resp := getStreamHTTP(t, client, baseURL, "sse-lines")
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read SSE response: %v", err)
	}

	want := "data: line one\ndata: line two\n\nevent: done\ndata: [DONE]\n\n"
	if got := string(body); got != want {
		t.Errorf("SSE body = %q, want %q", got, want)
	}
}

func TestReplacingChatCancelsPreviousRunAndKeepsLatestStream(t *testing.T) {
	firstStarted := make(chan struct{})
	var calls int32
	s := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			close(firstStarted)
			select {
			case <-r.Context().Done():
			case <-time.After(250 * time.Millisecond):
			}
			return
		}
		writeLLMStream(w, "latest")
	})
	client, baseURL := serveFiberApp(t, s.app)

	firstResp := postChatHTTP(t, client, baseURL, "replace-me", "first")
	firstResp.Body.Close()
	if firstResp.StatusCode != http.StatusAccepted {
		t.Fatalf("first /chat status = %d, want %d", firstResp.StatusCode, http.StatusAccepted)
	}
	select {
	case <-firstStarted:
	case <-time.After(time.Second):
		t.Fatal("first LLM request did not start")
	}
	s.mu.Lock()
	firstStream := s.channels["replace-me"]
	s.mu.Unlock()
	if firstStream == nil {
		t.Fatal("first active stream was not registered")
	}

	secondResp := postChatHTTP(t, client, baseURL, "replace-me", "second")
	secondResp.Body.Close()
	if secondResp.StatusCode != http.StatusAccepted {
		t.Fatalf("second /chat status = %d, want %d", secondResp.StatusCode, http.StatusAccepted)
	}

	select {
	case <-firstStream.finished:
	case <-time.After(time.Second):
		t.Fatal("replaced stream did not finish")
	}

	streamResp := getStreamHTTP(t, client, baseURL, "replace-me")
	defer streamResp.Body.Close()
	body, err := io.ReadAll(streamResp.Body)
	if err != nil {
		t.Fatalf("read replacement SSE response: %v", err)
	}
	if !bytes.Contains(body, []byte("data: latest\n\n")) {
		t.Errorf("replacement SSE stream = %q, want latest response", body)
	}
	if bytes.Contains(body, []byte("data: first")) {
		t.Errorf("replacement SSE stream contains stale response: %q", body)
	}
}
