package session

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedisStore(t *testing.T) (*RedisStore, *miniredis.Miniredis) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close Redis client: %v", err)
		}
	})
	store, err := NewRedisStore(client, "")
	if err != nil {
		t.Fatalf("NewRedisStore() returned error: %v", err)
	}
	return store, server
}

func TestRedisStoreCreateGetAndDelete(t *testing.T) {
	store, server := newTestRedisStore(t)
	ctx := context.Background()

	id, err := store.Create(ctx, "user-1", time.Hour)
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}
	if id == "" {
		t.Fatal("Create() returned empty token")
	}

	got, err := store.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}
	if got.UserID != "user-1" {
		t.Errorf("Get() user ID = %q, want user-1", got.UserID)
	}

	if server.TTL(defaultKeyPrefix+keyDigest(id)) != time.Hour {
		t.Errorf("stored session TTL = %v, want %v", server.TTL(defaultKeyPrefix+keyDigest(id)), time.Hour)
	}

	if err := store.Delete(ctx, id); err != nil {
		t.Fatalf("Delete() returned error: %v", err)
	}
	if _, err := store.Get(ctx, id); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get() after Delete() error = %v, want %v", err, ErrNotFound)
	}
}

func TestRedisStoreExpiresSession(t *testing.T) {
	store, server := newTestRedisStore(t)
	ctx := context.Background()

	id, err := store.Create(ctx, "user-1", time.Minute)
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}
	server.FastForward(time.Minute)

	if _, err := store.Get(ctx, id); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get() after expiration error = %v, want %v", err, ErrNotFound)
	}
}

func TestRedisStoreRejectsInvalidCreateArguments(t *testing.T) {
	store, _ := newTestRedisStore(t)
	for _, tt := range []struct {
		name string
		user string
		ttl  time.Duration
	}{
		{name: "empty user", ttl: time.Minute},
		{name: "zero ttl", user: "user-1"},
		{name: "negative ttl", user: "user-1", ttl: -time.Second},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if id, err := store.Create(context.Background(), tt.user, tt.ttl); !errors.Is(err, ErrInvalidSession) || id != "" {
				t.Errorf("Create() = (%q, %v), want empty ID and %v", id, err, ErrInvalidSession)
			}
		})
	}
}

func TestRedisStoreMissingAndEmptyTokens(t *testing.T) {
	store, _ := newTestRedisStore(t)

	if _, err := store.Get(context.Background(), "unknown"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get(unknown) error = %v, want %v", err, ErrNotFound)
	}
	if _, err := store.Get(context.Background(), ""); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get(empty) error = %v, want %v", err, ErrNotFound)
	}
	if err := store.Delete(context.Background(), ""); err != nil {
		t.Errorf("Delete(empty) error = %v, want nil", err)
	}
}

func keyDigest(id string) string {
	digest := sha256.Sum256([]byte(id))
	return hex.EncodeToString(digest[:])
}
