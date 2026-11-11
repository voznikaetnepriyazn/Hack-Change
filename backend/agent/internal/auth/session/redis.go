package session

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultKeyPrefix  = "auth:session:"
	sessionIDAttempts = 3
)

// RedisConfig contains connection settings for the Redis session store.
type RedisConfig struct {
	Addr      string
	Password  string
	DB        int
	KeyPrefix string
}

// RedisStore stores user identity by a hash of the opaque session token.
// Redis expires each entry automatically after the TTL supplied to Create.
type RedisStore struct {
	client     redis.UniversalClient
	keyPrefix  string
	ownsClient bool
}

var _ Store = (*RedisStore)(nil)

// NewRedisStore creates a store using an existing Redis client.
func NewRedisStore(client redis.UniversalClient, keyPrefix string) (*RedisStore, error) {
	if client == nil {
		return nil, ErrInvalidStoreConfig
	}
	keyPrefix = strings.TrimSpace(keyPrefix)
	if keyPrefix == "" {
		keyPrefix = defaultKeyPrefix
	}
	return &RedisStore{client: client, keyPrefix: keyPrefix}, nil
}

// OpenRedisStore connects to Redis and verifies connectivity before returning.
func OpenRedisStore(ctx context.Context, cfg RedisConfig) (*RedisStore, error) {
	if strings.TrimSpace(cfg.Addr) == "" || cfg.DB < 0 {
		return nil, ErrInvalidStoreConfig
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	store, err := NewRedisStore(client, cfg.KeyPrefix)
	if err != nil {
		_ = client.Close()
		return nil, err
	}
	store.ownsClient = true
	return store, nil
}

// Create allocates a token and stores its user ID with the requested expiration.
func (s *RedisStore) Create(ctx context.Context, userID string, ttl time.Duration) (string, error) {
	if strings.TrimSpace(userID) == "" || ttl <= 0 {
		return "", ErrInvalidSession
	}

	for range sessionIDAttempts {
		id, err := NewID()
		if err != nil {
			return "", err
		}
		created, err := s.client.SetNX(ctx, s.key(id), userID, ttl).Result()
		if err != nil {
			return "", err
		}
		if created {
			return id, nil
		}
	}
	return "", ErrIDCollision
}

// Get resolves a token to its authenticated user ID.
func (s *RedisStore) Get(ctx context.Context, id string) (Session, error) {
	if id == "" {
		return Session{}, ErrNotFound
	}
	userID, err := s.client.Get(ctx, s.key(id)).Result()
	if errors.Is(err, redis.Nil) {
		return Session{}, ErrNotFound
	}
	if err != nil {
		return Session{}, err
	}
	if userID == "" {
		return Session{}, ErrNotFound
	}
	return Session{UserID: userID}, nil
}

// Delete revokes a token. Deleting a missing token is not an error.
func (s *RedisStore) Delete(ctx context.Context, id string) error {
	if id == "" {
		return nil
	}
	return s.client.Del(ctx, s.key(id)).Err()
}

// Close closes the Redis client owned by this store.
func (s *RedisStore) Close() error {
	if !s.ownsClient {
		return nil
	}
	return s.client.Close()
}

func (s *RedisStore) key(id string) string {
	digest := sha256.Sum256([]byte(id))
	return s.keyPrefix + hex.EncodeToString(digest[:])
}
