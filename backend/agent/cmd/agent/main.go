package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"

	agent "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/agent"
	authhandler "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/handler"
	authmiddleware "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/middleware"
	authservice "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/service"
	authsession "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/session"
	authstorage "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/storage"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/httpclient"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/server"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/rag/embed"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/rag/storage"
	tool "github.com/voznikaetnepriyazn/Hack-Change/agent/tools"
	answerquestions "github.com/voznikaetnepriyazn/Hack-Change/agent/tools/answer_questions"
)

type runtimeConfig struct {
	redisAddr      string
	redisDB        int
	redisKeyPrefix string
	usersFile      string
	sessionTTL     time.Duration
	cookieName     string
	cookieSecure   bool
	cookieSameSite http.SameSite
	allowedOrigins []string
}

const serverPort = "8080"

func main() {
	if err := run(); err != nil {
		slog.Error("Agent startup failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("load environment file: %w", err)
	}

	SetupLogger("agent.log", slog.LevelDebug)

	modelID := strings.TrimSpace(os.Getenv("MODEL_ID"))
	if modelID == "" {
		return errors.New("MODEL_ID is required")
	}
	token := strings.TrimSpace(os.Getenv("API_KEY"))
	if token == "" {
		return errors.New("API_KEY is required")
	}

	cfg, err := loadRuntimeConfig(os.Getenv)
	if err != nil {
		return err
	}
	httpClient, err := newProxyHTTPClient(os.Getenv("HTTPS_PROXY"), os.Getenv("HTTP_PROXY"))
	if err != nil {
		return fmt.Errorf("configure LLM HTTP client: %w", err)
	}

	llmConfig := openai.DefaultConfig(token)
	llmConfig.HTTPClient = httpClient
	if baseURL := strings.TrimSpace(os.Getenv("BASE_URL")); baseURL != "" {
		llmConfig.BaseURL = baseURL
	}
	client := openai.NewClientWithConfig(llmConfig)

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required for answer_questions")
	}
	embeddingAPIKey := strings.TrimSpace(os.Getenv("EMBED_KEY"))
	if embeddingAPIKey == "" {
		return errors.New("EMBED_KEY is required for answer_questions")
	}
	ragPool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return fmt.Errorf("configure RAG PostgreSQL connection: %w", err)
	}
	defer ragPool.Close()

	ragConnectCtx, cancelRAGConnect := context.WithTimeout(context.Background(), 10*time.Second)
	err = ragPool.Ping(ragConnectCtx)
	cancelRAGConnect()
	if err != nil {
		return fmt.Errorf("connect to RAG PostgreSQL database: %w", err)
	}
	answerTool := answerquestions.NewAnswerTool(
		embed.NewEmbedderWithHTTPClient(
			embeddingAPIKey,
			os.Getenv("EMBED_URL"),
			os.Getenv("EMBED_MODEL"),
			httpClient,
		),
		storage.NewStore(ragPool),
	)

	registry := tool.NewRegistry()
	registry.RegisterAll(answerTool)
	slog.Info("Registered agent tools", "count", len(registry.Schemas()))

	assistant := agent.NewAgent(client, registry, modelID, strings.EqualFold(os.Getenv("LLM_STREAM"), "true"))

	users, err := authstorage.NewReading(cfg.usersFile)
	if err != nil {
		return fmt.Errorf("load authentication users: %w", err)
	}
	auth, err := authservice.New(users)
	if err != nil {
		return fmt.Errorf("initialize authentication service: %w", err)
	}

	connectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	sessions, err := authsession.OpenRedisStore(connectCtx, authsession.RedisConfig{
		Addr:      cfg.redisAddr,
		DB:        cfg.redisDB,
		KeyPrefix: cfg.redisKeyPrefix,
	})
	cancel()
	if err != nil {
		return fmt.Errorf("connect to Redis session store: %w", err)
	}
	defer func() {
		if err := sessions.Close(); err != nil {
			slog.Error("Close Redis session store failed", "error", err)
		}
	}()

	authHandler, err := authhandler.New(auth, sessions, authhandler.Config{
		CookieName: cfg.cookieName,
		SessionTTL: cfg.sessionTTL,
		Secure:     cfg.cookieSecure,
		SameSite:   cfg.cookieSameSite,
	})
	if err != nil {
		return fmt.Errorf("initialize authentication handler: %w", err)
	}
	authMiddleware, err := authmiddleware.New(sessions, auth, authmiddleware.Config{
		CookieName: cfg.cookieName,
	})
	if err != nil {
		return fmt.Errorf("initialize authentication middleware: %w", err)
	}

	srv, err := server.NewServerWithAuth(assistant, server.AuthConfig{
		Handler:        authHandler,
		Middleware:     authMiddleware,
		AllowedOrigins: cfg.allowedOrigins,
	})
	if err != nil {
		return fmt.Errorf("initialize authenticated server: %w", err)
	}

	slog.Info("Authenticated agent server starting", "port", serverPort)
	if err := srv.Start(serverPort); err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	return nil
}

func loadRuntimeConfig(getenv func(string) string) (runtimeConfig, error) {
	cfg := runtimeConfig{
		redisAddr:      strings.TrimSpace(getenv("REDIS_ADDR")),
		redisKeyPrefix: strings.TrimSpace(getenv("REDIS_KEY_PREFIX")),
		usersFile:      strings.TrimSpace(getenv("AUTH_USERS_FILE")),
		cookieName:     strings.TrimSpace(getenv("AUTH_COOKIE_NAME")),
	}
	if cfg.redisAddr == "" {
		cfg.redisAddr = "host.docker.internal:6379"
	}
	if cfg.redisKeyPrefix == "" {
		cfg.redisKeyPrefix = "auth:session:"
	}
	if cfg.usersFile == "" {
		cfg.usersFile = filepath.Join("agent", "internal", "auth", "users.json")
	}
	if cfg.cookieName == "" {
		cfg.cookieName = "auth_session"
	}
	cfg.redisDB = 0
	var err error
	if rawDB := strings.TrimSpace(getenv("REDIS_DB")); rawDB != "" {
		cfg.redisDB, err = strconv.Atoi(rawDB)
		if err != nil || cfg.redisDB < 0 {
			return runtimeConfig{}, errors.New("REDIS_DB must be a non-negative integer")
		}
	}

	cfg.sessionTTL = 24 * time.Hour
	if rawTTL := strings.TrimSpace(getenv("AUTH_SESSION_TTL")); rawTTL != "" {
		cfg.sessionTTL, err = time.ParseDuration(rawTTL)
		if err != nil || cfg.sessionTTL <= 0 {
			return runtimeConfig{}, errors.New("AUTH_SESSION_TTL must be a positive Go duration")
		}
	}

	cfg.cookieSecure = false
	if rawSecure := strings.TrimSpace(getenv("AUTH_COOKIE_SECURE")); rawSecure != "" {
		cfg.cookieSecure, err = strconv.ParseBool(rawSecure)
		if err != nil {
			return runtimeConfig{}, errors.New("AUTH_COOKIE_SECURE must be true or false")
		}
	}
	if strings.EqualFold(strings.TrimSpace(getenv("APP_ENV")), "production") && !cfg.cookieSecure {
		return runtimeConfig{}, errors.New("AUTH_COOKIE_SECURE must be true in production")
	}
	switch strings.ToLower(strings.TrimSpace(getenv("AUTH_COOKIE_SAME_SITE"))) {
	case "", "lax":
		cfg.cookieSameSite = http.SameSiteLaxMode
	case "strict":
		cfg.cookieSameSite = http.SameSiteStrictMode
	case "none":
		cfg.cookieSameSite = http.SameSiteNoneMode
		if !cfg.cookieSecure {
			return runtimeConfig{}, errors.New("AUTH_COOKIE_SECURE must be true when AUTH_COOKIE_SAME_SITE=none")
		}
	default:
		return runtimeConfig{}, errors.New("AUTH_COOKIE_SAME_SITE must be lax, strict, or none")
	}

	origins := strings.TrimSpace(getenv("CORS_ALLOWED_ORIGINS"))
	if origins == "" {
		origins = "http://localhost:3000"
	}
	for _, origin := range strings.Split(origins, ",") {
		origin = strings.TrimSpace(origin)
		if origin == "" || origin == "*" {
			return runtimeConfig{}, errors.New("CORS_ALLOWED_ORIGINS must contain explicit origins, not an empty value or wildcard")
		}
		cfg.allowedOrigins = append(cfg.allowedOrigins, origin)
	}
	return cfg, nil
}

func newProxyHTTPClient(httpsProxy, httpProxy string) (*http.Client, error) {
	return httpclient.NewSOCKS5ProxyClient(httpsProxy, httpProxy)
}

func SetupLogger(logFileName string, level slog.Level) {
	logDir := os.Getenv("LOG_PATH")
	if logDir == "" {
		logDir = "logs"
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic("can not create logs dir: " + err.Error())
	}

	path := filepath.Join(logDir, logFileName)

	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic("can not open log file: " + err.Error())
	}

	writer := io.MultiWriter(os.Stderr, logFile)

	logger := slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				t := a.Value.Time()
				return slog.Attr{
					Key:   slog.TimeKey,
					Value: slog.StringValue(t.Format("2006-01-02 15:04:05")),
				}
			}
			return a
		},
	}))

	slog.SetDefault(logger)

	slog.Info("Started logger",
		"file", path,
		"level", level.String(),
		"current time", time.Now().Format("15:04:05"))
}
