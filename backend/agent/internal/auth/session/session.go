package session

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound           = errors.New("session not found")
	ErrInvalidSession     = errors.New("invalid session")
	ErrIDCollision        = errors.New("could not allocate unique session ID")
	ErrInvalidStoreConfig = errors.New("invalid Redis session store configuration")
)

// Session contains the authenticated identity associated with a session token.
type Session struct {
	UserID string
}

// Store persists authentication sessions with an expiration.
type Store interface {
	Create(ctx context.Context, userID string, ttl time.Duration) (string, error)
	Get(ctx context.Context, id string) (Session, error)
	Delete(ctx context.Context, id string) error
}
