package main

import (
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func env(values map[string]string) func(string) string {
	return func(key string) string {
		return values[key]
	}
}

func TestLoadRuntimeConfigDefaultsAndExplicitOrigins(t *testing.T) {
	cfg, err := loadRuntimeConfig(env(map[string]string{}))
	if err != nil {
		t.Fatalf("loadRuntimeConfig() error = %v", err)
	}

	if cfg.redisAddr != "host.docker.internal:6379" || cfg.redisDB != 0 {
		t.Errorf("Redis config = %q DB %d, want host.docker.internal:6379 DB 0", cfg.redisAddr, cfg.redisDB)
	}
	if cfg.usersFile != filepath.Join("agent", "internal", "auth", "users.json") {
		t.Errorf("usersFile = %q, want default auth users file", cfg.usersFile)
	}
	if cfg.sessionTTL != 24*time.Hour {
		t.Errorf("sessionTTL = %v, want 24h", cfg.sessionTTL)
	}
	if cfg.cookieName != "auth_session" || cfg.cookieSecure || cfg.cookieSameSite != http.SameSiteLaxMode {
		t.Errorf("cookie config = %q secure=%v sameSite=%v, want auth_session/false/Lax", cfg.cookieName, cfg.cookieSecure, cfg.cookieSameSite)
	}
	if serverPort != "8080" {
		t.Errorf("serverPort = %q, want 8080", serverPort)
	}
	if !reflect.DeepEqual(cfg.allowedOrigins, []string{"http://localhost:3000"}) {
		t.Errorf("allowedOrigins = %v, want local development origin", cfg.allowedOrigins)
	}
}

func TestLoadRuntimeConfigParsesAuthAndServerSettings(t *testing.T) {
	cfg, err := loadRuntimeConfig(env(map[string]string{
		"REDIS_ADDR":            "redis.internal:6380",
		"REDIS_DB":              "2",
		"REDIS_KEY_PREFIX":      "app:sessions:",
		"AUTH_USERS_FILE":       "config/users.json",
		"AUTH_SESSION_TTL":      "12h",
		"AUTH_COOKIE_NAME":      "session",
		"AUTH_COOKIE_SECURE":    "true",
		"AUTH_COOKIE_SAME_SITE": "none",
		"CORS_ALLOWED_ORIGINS":  "https://app.example, https://admin.example",
	}))
	if err != nil {
		t.Fatalf("loadRuntimeConfig() error = %v", err)
	}
	if cfg.redisDB != 2 || cfg.redisAddr != "redis.internal:6380" || cfg.redisKeyPrefix != "app:sessions:" {
		t.Errorf("Redis settings were not parsed: %+v", cfg)
	}
	if cfg.usersFile != "config/users.json" || cfg.sessionTTL != 12*time.Hour {
		t.Errorf("auth file/TTL = %q/%v", cfg.usersFile, cfg.sessionTTL)
	}
	if cfg.cookieName != "session" || !cfg.cookieSecure || cfg.cookieSameSite != http.SameSiteNoneMode {
		t.Errorf("cookie config = %q secure=%v sameSite=%v", cfg.cookieName, cfg.cookieSecure, cfg.cookieSameSite)
	}
	if !reflect.DeepEqual(cfg.allowedOrigins, []string{"https://app.example", "https://admin.example"}) {
		t.Errorf("server origins = %v", cfg.allowedOrigins)
	}
}

func TestLoadRuntimeConfigRejectsInvalidSettings(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]string
		want   string
	}{
		{name: "invalid Redis database", values: map[string]string{"REDIS_DB": "-1"}, want: "REDIS_DB"},
		{name: "invalid session TTL", values: map[string]string{"AUTH_SESSION_TTL": "0s"}, want: "AUTH_SESSION_TTL"},
		{name: "invalid secure flag", values: map[string]string{"AUTH_COOKIE_SECURE": "sometimes"}, want: "AUTH_COOKIE_SECURE"},
		{
			name: "SameSite none requires secure cookie",
			values: map[string]string{
				"AUTH_COOKIE_SAME_SITE": "none",
			},
			want: "AUTH_COOKIE_SECURE",
		},
		{
			name: "production requires secure cookie",
			values: map[string]string{
				"APP_ENV": "production",
			},
			want: "AUTH_COOKIE_SECURE",
		},
		{name: "wildcard CORS origin", values: map[string]string{"CORS_ALLOWED_ORIGINS": "*"}, want: "CORS_ALLOWED_ORIGINS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := map[string]string{}
			for key, value := range tt.values {
				values[key] = value
			}
			_, err := loadRuntimeConfig(env(values))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("loadRuntimeConfig() error = %v, want error containing %q", err, tt.want)
			}
		})
	}
}

func TestNewProxyHTTPClient(t *testing.T) {
	tests := []struct {
		name       string
		httpsProxy string
		httpProxy  string
		wantErr    bool
	}{
		{name: "SOCKS5 proxy", httpsProxy: "socks5://proxy.example:1080"},
		{name: "HTTP proxy variable fallback", httpProxy: "socks5://proxy.example:1080"},
		{name: "SOCKS5 authentication", httpsProxy: "socks5://user:pass@proxy.example:1080"},
		{name: "proxy required", wantErr: true},
		{name: "invalid proxy", httpsProxy: "://bad", wantErr: true},
		{name: "reject HTTP proxy", httpsProxy: "http://proxy.example:8080", wantErr: true},
		{name: "reject SOCKS5h scheme", httpsProxy: "socks5h://proxy.example:1080", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := newProxyHTTPClient(tt.httpsProxy, tt.httpProxy)
			if (err != nil) != tt.wantErr {
				t.Fatalf("newProxyHTTPClient() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				if client == nil || client.Timeout != 180*time.Second || client.Transport == nil {
					t.Errorf("client = %+v, want configured HTTP client with 180s timeout", client)
				}
				transport, ok := client.Transport.(*http.Transport)
				if !ok || transport.DialContext == nil || transport.Proxy != nil {
					t.Errorf("transport = %#v, want SOCKS5 DialContext and no HTTP proxy", client.Transport)
				}
			}
		})
	}
}

func TestNewProxyHTTPClientRejectsMalformedSOCKS5URL(t *testing.T) {
	for _, value := range []string{"socks5://", "socks5://proxy.example", "socks5://proxy.example:0"} {
		t.Run(value, func(t *testing.T) {
			_, err := newProxyHTTPClient(value, "")
			if err == nil {
				t.Errorf("newProxyHTTPClient(%q) error = nil, want invalid SOCKS5 URL", value)
			}
		})
	}
}
